class StorageError(RuntimeError):
    """Base exception for storage related failures."""


class ObjectNotFound(StorageError):
    """Raised when an object does not exist in storage."""


class CatalogError(RuntimeError):
    """Base exception for catalog related failures."""


class TrackNotFound(CatalogError):
    """Raised when a track cannot be located."""

