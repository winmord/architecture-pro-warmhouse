import asyncio
import json
import logging
import os
from contextlib import asynccontextmanager
from datetime import datetime
from typing import List, Optional

import aio_pika
from fastapi import FastAPI, Query, HTTPException

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class Storage:
    def __init__(self):
        self.data = []
        self.next_id = 1

    def add(self, device_id: str, reading_type: str, value: float, unit: str = ""):
        item = {
            "id": self.next_id,
            "device_id": device_id,
            "type": reading_type,
            "value": value,
            "unit": unit,
            "time": datetime.now().isoformat()
        }
        self.data.append(item)
        self.next_id += 1
        return item

    def get_by_device(self, device_id: str, reading_type: Optional[str] = None, limit: int = 1000):
        result = []
        for item in self.data:
            if item["device_id"] == device_id:
                if reading_type and item["type"] != reading_type:
                    continue
                result.append(item)
        result.sort(key=lambda x: x["time"], reverse=True)
        return result[:limit]

    def get_stats(self, device_ids: List[str], reading_type: str):
        values = []
        for item in self.data:
            if item["device_id"] in device_ids and item["type"] == reading_type:
                values.append(item["value"])

        if not values:
            return {"count": 0}

        return {
            "avg": sum(values) / len(values),
            "min": min(values),
            "max": max(values),
            "count": len(values)
        }


storage = Storage()


class RabbitConsumer:
    def __init__(self):
        self.running = False

    async def start(self):
        try:
            rabbitmq_url = f"amqp://user:pass@rabbitmq:5672/"

            connection = await aio_pika.connect_robust(rabbitmq_url)
            channel = await connection.channel()

            exchange = await channel.declare_exchange("device.exchange", aio_pika.ExchangeType.TOPIC, durable=True)
            queue = await channel.declare_queue("telemetry.queue")
            await queue.bind(exchange, routing_key="device.telemetry")

            print("Connected to RabbitMQ")

            async with queue.iterator() as queue_iter:
                async for message in queue_iter:
                    async with message.process():
                        body = message.body.decode()
                        await self.handle_message(body)

        except Exception as e:
            print(f"RabbitMQ error: {e}")

    @staticmethod
    async def handle_message(body: str):
        try:
            data = json.loads(body)
            device_id = data.get("device_id")
            telemetry_type = data.get("type")
            unit = data.get("unit")
            value = data.get("value")

            if value is not None:
                storage.add(device_id, telemetry_type, float(value), unit)

            print(f"Saved telemetry readings for device {device_id}")

        except Exception as e:
            print(f"Message error: {e}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    print("Starting Telemetry Service...")

    consumer = RabbitConsumer()
    asyncio.create_task(consumer.start())

    yield

    print("Shutting down...")


app = FastAPI(lifespan=lifespan)


@app.get("/api/v1/telemetry/{device_id}")
async def get_telemetry(
        device_id: str,
        type: Optional[str] = Query(None, alias="reading_type"),
        limit: int = Query(1000, ge=1, le=10000)
):
    logger.info(f"GET /telemetry/{device_id} - type: {type}, limit: {limit}")
    data = storage.get_by_device(device_id, type, limit)
    if not data:
        raise HTTPException(status_code=404, detail="No data found")
    return data


@app.get("/health")
async def health():
    logger.info(f"GET /health/")
    return {
        "status": "healthy",
        "data_count": len(storage.data),
        "service": "telemetry-service"
    }


@app.get("/")
async def root():
    return {
        "service": "Telemetry Service",
        "endpoints": {
            "GET /telemetry/{device_id}": "Get device telemetry",
            "GET /telemetry/stats?device_ids=id1,id2&reading_type=temperature": "Get statistics",
            "GET /health": "Health check"
        },
        "rabbitmq": {
            "exchange": "device.exchange",
            "routing_key": "device.telemetry",
            "queue": "telemetry.queue"
        }
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8083,
        log_level="info",
        access_log=True
    )
