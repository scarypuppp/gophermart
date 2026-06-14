"""
Mock accrual service for local development.

Orders are auto-registered on first query and processed after 10 seconds.
50% chance of INVALID, 50% chance of PROCESSED with random accrual.

Usage:
    pip install fastapi uvicorn
    python scripts/accrual_mock.py
"""

import asyncio
import random

import uvicorn
from fastapi import FastAPI, Response

app = FastAPI()

# order_number -> {"status": str, "accrual": float | None}
orders: dict[str, dict] = {}


async def _process(number: str) -> None:
    await asyncio.sleep(10)
    if random.random() < 0.5:
        orders[number] = {"status": "INVALID"}
    else:
        orders[number] = {
            "status": "PROCESSED",
            "accrual": round(random.uniform(10, 1000), 2),
        }
    print(f"[accrual] {number} -> {orders[number]}")


@app.get("/api/orders/{number}")
async def get_order(number: str, response: Response):
    if number not in orders:
        orders[number] = {"status": "PROCESSING"}
        asyncio.create_task(_process(number))
        print(f"[accrual] registered {number}, processing...")

    data = orders[number]
    body: dict = {"order": number, "status": data["status"]}
    if "accrual" in data:
        body["accrual"] = data["accrual"]

    return body


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8081)
