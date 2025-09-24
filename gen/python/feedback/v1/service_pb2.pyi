from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class FeedbackType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FEEDBACK_TYPE_UNSPECIFIED: _ClassVar[FeedbackType]
    FEEDBACK_TYPE_GOOD: _ClassVar[FeedbackType]
    FEEDBACK_TYPE_BAD: _ClassVar[FeedbackType]
FEEDBACK_TYPE_UNSPECIFIED: FeedbackType
FEEDBACK_TYPE_GOOD: FeedbackType
FEEDBACK_TYPE_BAD: FeedbackType

class SubmitFeedbackRequest(_message.Message):
    __slots__ = ("type", "thread")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    THREAD_FIELD_NUMBER: _ClassVar[int]
    type: FeedbackType
    thread: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, type: _Optional[_Union[FeedbackType, str]] = ..., thread: _Optional[_Iterable[str]] = ...) -> None: ...

class SubmitFeedbackResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
