# server/services/killswitch_service.py

import asyncio
from datetime import datetime, timedelta
import sys

class KillSwitchService:
    def __init__(self):
        self._shutdown_event = asyncio.Event()

    def trigger_shutdown(self):
        """Manual termination only"""
        self._shutdown_event.set()
        print("+-----------------------------------------+")
        print("| [KillSwitch] Manual shutdown triggered. |")
        print("+-----------------------------------------+")

    async def wait_for_shutdown(self):
        await self._shutdown_event.wait()

    async def manual_terminate_listener(self):
        while not self._shutdown_event.is_set():
            try:
                cmd = await asyncio.to_thread(input, "")
                if cmd.strip().lower() == "terminate":
                    self.trigger_shutdown()
                    break
            except Exception:
                continue