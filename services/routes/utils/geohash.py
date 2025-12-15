"""Geohash utilities for location-based queries."""

import geohash2


def encode_geohash(latitude: float, longitude: float, precision: int = 9) -> str:
    """
    Encode latitude and longitude to geohash.

    Args:
        latitude: Latitude in degrees
        longitude: Longitude in degrees
        precision: Geohash precision (default 9, ~5 meters)

    Returns:
        Geohash string
    """
    return geohash2.encode(latitude, longitude, precision=precision)


def decode_geohash(geohash: str) -> tuple[float, float]:
    """
    Decode geohash to latitude and longitude.

    Args:
        geohash: Geohash string

    Returns:
        Tuple of (latitude, longitude)
    """
    return geohash2.decode(geohash)


def get_neighbors(geohash: str) -> list[str]:
    """
    Get neighboring geohashes.

    Args:
        geohash: Geohash string

    Returns:
        List of neighboring geohashes (including the original)
    """
    return geohash2.get_neighbors(geohash)


def get_neighbors_with_self(geohash: str) -> list[str]:
    """
    Get neighboring geohashes including the original.

    Args:
        geohash: Geohash string

    Returns:
        List of neighboring geohashes including the original
    """
    neighbors = get_neighbors(geohash)
    return [geohash] + neighbors

