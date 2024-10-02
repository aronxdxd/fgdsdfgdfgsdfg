import asyncio
import random
import string
from aiogram import Bot, Dispatcher, types
from aiogram.filters.command import Command
from aiogram.filters import CommandObject

from motor.motor_asyncio import AsyncIOMotorClient
import json


def load_config():
    try:
        with open("config.json", "r") as config_file:
            return json.load(config_file)
    except FileNotFoundError:
        print("Error: config.json file not found.")
        return None
    except json.JSONDecodeError:
        print("Error: Invalid JSON in config.json file.")
        return None


config = load_config()
BOT_TOKEN = config.get("token") if config else "YOUR_BOT_TOKEN"

bot = Bot(token=BOT_TOKEN)
dp = Dispatcher()

mongo_client = AsyncIOMotorClient("mongodb://localhost:27017")
db = mongo_client.db
users_collection = db.users


@dp.message(Command("start"))
async def cmd_start(message: types.Message, command: CommandObject):
    user_id = message.from_user.id
    referral_code = "".join(random.choices(string.ascii_letters + string.digits, k=6))

    referrer_code = command.args
    referrer = None
    if referrer_code:
        referrer = await users_collection.find_one({"referral_code": referrer_code})

    user_data = {
        "user_id": user_id,
        "first_name": message.from_user.first_name,
        "joined_at": message.date,
        "balance": 0,
        "referrals": [],
        "referral_code": referral_code,
        "energy": 500,
        "energy_max": 500,
        "referred_by": referrer["user_id"] if referrer else None,
        "modifies": {
            "toques_lvl": 1,
            "subscribe_bonus": False,
        },
    }

    if not await users_collection.count_documents({"user_id": user_id}):
        await users_collection.insert_one(user_data)
        if referrer:
            await users_collection.update_one(
                {"user_id": referrer["user_id"]},
                {"$push": {"referrals": user_id}, "$inc": {"energy_max": 100}},
            )

    welcome_message = f"Welcome!"

    await message.answer(
        welcome_message,
        reply_markup=types.InlineKeyboardMarkup(
            inline_keyboard=[
                [
                    types.InlineKeyboardButton(
                        text="Open Web App",
                        web_app=types.WebAppInfo(
                            url="https://tgwebapp-two.vercel.app/"
                        ),
                    )
                ]
            ]
        ),
    )


async def main():
    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
