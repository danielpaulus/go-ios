from enum import Enum


class JobStatusType1(str, Enum):
    FAILED = "failed"
    RUNNING = "running"
    STOPPED = "stopped"
    SUCCEEDED = "succeeded"

    def __str__(self) -> str:
        return str(self.value)
