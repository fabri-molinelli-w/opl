"""Historial de conversación en memoria del proceso (CU-09 RN-3: últimos N turnos por sesión)."""

from collections import OrderedDict, deque

from langchain_core.messages import AIMessage, BaseMessage, HumanMessage

MAX_SESSIONS = 500


class SessionMemory:
    def __init__(self, turns: int) -> None:
        self._turns = turns
        self._sessions: OrderedDict[str, deque[BaseMessage]] = OrderedDict()

    @staticmethod
    def key(user_id: str, session_id: str) -> str:
        # El usuario forma parte de la clave: una sesión nunca se comparte entre usuarios.
        return f"{user_id}:{session_id}"

    def history(self, key: str) -> list[BaseMessage]:
        return list(self._sessions.get(key, ()))

    def append_turn(self, key: str, question: str, answer: str) -> None:
        messages = self._sessions.get(key)
        if messages is None:
            messages = deque(maxlen=self._turns * 2)
            self._sessions[key] = messages
        self._sessions.move_to_end(key)
        messages.append(HumanMessage(content=question))
        messages.append(AIMessage(content=answer))
        while len(self._sessions) > MAX_SESSIONS:
            self._sessions.popitem(last=False)
