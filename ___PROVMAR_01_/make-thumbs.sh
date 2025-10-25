#!/usr/bin/env bash
# Generate thumbnails (JPEG) for mp4 files in ./video
# Requires ffmpeg installed on the machine

set -euo pipefail

VIDEO_DIR="video"
OUT_DIR="media/thumbs"
mkdir -p "$OUT_DIR"

for f in "$VIDEO_DIR"/*.mp4; do
  [ -e "$f" ] || continue
  base=$(basename "$f" .mp4)
  out="$OUT_DIR/${base}.jpg"
  if [ -f "$out" ]; then
    echo "Skipping existing thumbnail: $out"
    continue
  fi
  echo "Generating thumbnail for $f -> $out"
  ffmpeg -y -ss 3 -i "$f" -vframes 1 -q:v 2 -vf "scale=640:-1" "$out" || echo "ffmpeg failed for $f"
done

echo "Done. Thumbnails in $OUT_DIR"
