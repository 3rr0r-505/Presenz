# server/services/killswitch_service.py

import asyncio
from datetime import datetime, timedelta


class KillSwitchService:
    def __init__(self, timeout_minutes: int = 3) -> None:
        # Shared shutdown event for both inactivity and manual terminate
        self._shutdown_event = asyncio.Event()
        self._last_activity = datetime.utcnow()
        self._timeout = timedelta(minutes=timeout_minutes)

    # ----------------------------
    # Manual / shared shutdown
    # ----------------------------
    def trigger_shutdown(self) -> None:
        """Trigger shutdown (used by both inactivity and manual terminate)."""
        if not self._shutdown_event.is_set():
            self._shutdown_event.set()
            print("+-----------------------------------------+")
            print("| [KillSwitch] Shutdown triggered.        |")
            print("+-----------------------------------------+")

    async def wait_for_shutdown(self) -> None:
        """Wait until any shutdown path fires."""
        await self._shutdown_event.wait()

    async def manual_terminate_listener(self) -> None:
        """
        Wait for user to type 'terminate' in the terminal.
        Works even when inactivity monitor is running.
        """
        while not self._shutdown_event.is_set():
            try:
                cmd = await asyncio.to_thread(input, "")
                if cmd.strip().lower() == "terminate":
                    print("[KillSwitch] 'terminate' command received.")
                    self.trigger_shutdown()
                    break
            except Exception:
                # Ignore input errors and keep listening
                continue

    # ----------------------------
    # Inactivity tracking
    # ----------------------------
    def update_activity(self) -> None:
        """Update last activity timestamp on each HTTP request."""
        self._last_activity = datetime.utcnow()

    async def inactivity_monitor(self) -> None:
        """
        Periodically check for inactivity.
        If no activity for timeout_minutes, trigger shutdown.
        """
        print(f"[KillSwitch] Inactivity monitor started (timeout: {self._timeout}).")
        while not self._shutdown_event.is_set():
            await asyncio.sleep(5)
            if datetime.utcnow() - self._last_activity > self._timeout:
                print("[KillSwitch] Inactivity timeout reached. Shutting down.")
                self.trigger_shutdown()
                break