from enum import Enum


class FileDomainType1(str, Enum):
    APP = "app"
    APP_GROUP = "app-group"
    CRASH = "crash"
    TEMP = "temp"

    def __str__(self) -> str:
        return str(self.value)
