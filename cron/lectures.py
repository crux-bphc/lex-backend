from surrealdb import Surreal
import os

async def aggregate_lectures():
    """
    TODO: scrapes the lectures from all tokens and then compiles them to a single
    table where there are no duplicate lectures
    """
    async with Surreal("ws://db:8000/rpc") as db:
        await db.signin({"user": os.environ["DB_USER"], "pass": os.environ["DB_PASSWORD"]})
        await db.use(os.environ["DB_NAMESPACE"], os.environ["DB_DATABASE"])

if __name__ == "__main__":
    import asyncio
    asyncio.run(aggregate_lectures())