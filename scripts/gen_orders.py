#!/usr/bin/env python3
"""
Generate and submit orders with valid Luhn numbers.
Usage:
    python gen_orders.py [--count 20] [--base-url http://localhost:8080] \
                         [--login user] [--password pass]
"""

import argparse
import random
import requests
import sys


def luhn_checksum(number: str) -> int:
    digits = [int(d) for d in number]
    n = len(digits)
    parity = n % 2
    total = 0
    for i, digit in enumerate(digits):
        if i % 2 == parity:
            digit *= 2
            if digit > 9:
                digit -= 9
        total += digit
    return total % 10


def generate_luhn(length: int = 16) -> str:
    """Generate a random number of given length that passes Luhn check."""
    while True:
        digits = [random.randint(0, 9) for _ in range(length - 1)]
        prefix = "".join(map(str, digits))
        # Try each possible check digit
        for check in range(10):
            candidate = prefix + str(check)
            if luhn_checksum(candidate) == 0:
                return candidate


def register_or_login(base_url: str, login: str, password: str) -> str:
    """Register user (ignore 409), then login and return token."""
    requests.post(
        f"{base_url}/api/user/register",
        json={"login": login, "password": password},
    )
    resp = requests.post(
        f"{base_url}/api/user/login",
        json={"login": login, "password": password},
    )
    resp.raise_for_status()
    return resp.json()["access_token"]


def submit_order(base_url: str, token: str, number: str) -> tuple[int, str]:
    resp = requests.post(
        f"{base_url}/api/user/orders",
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "text/plain",
        },
        data=number,
    )
    return resp.status_code, number


def main():
    parser = argparse.ArgumentParser(description="Generate Luhn orders")
    parser.add_argument("--count", type=int, default=20, help="Number of orders to generate")
    parser.add_argument("--base-url", default="http://localhost:8080")
    parser.add_argument("--login", default="testuser")
    parser.add_argument("--password", default="testpass123")
    parser.add_argument("--length", type=int, default=16, help="Order number length")
    args = parser.parse_args()

    print(f"Logging in as '{args.login}'...")
    try:
        token = register_or_login(args.base_url, args.login, args.password)
    except Exception as e:
        print(f"Auth failed: {e}", file=sys.stderr)
        sys.exit(1)
    print("OK")

    ok = 0
    for i in range(args.count):
        number = generate_luhn(args.length)
        code, num = submit_order(args.base_url, token, number)
        status = {200: "already exists", 202: "accepted", 409: "conflict"}.get(code, f"error {code}")
        print(f"[{i+1:>3}/{args.count}] {num}  →  {status}")
        if code in (200, 202):
            ok += 1

    print(f"\nDone: {ok}/{args.count} submitted successfully.")


if __name__ == "__main__":
    main()
