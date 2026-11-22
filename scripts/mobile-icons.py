#!/usr/bin/env python3
"""Generate branded Android launcher and splash assets from the J3 logo.

Run automatically before every mobile APK build. Keeping this in source control
means a fresh checkout never falls back to the stock Capacitor icon.
"""
from pathlib import Path

from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parent.parent
LOGO = ROOT / "web/static/brand/logo.png"
RES = ROOT / "web/android/app/src/main/res"
NAVY = (3, 31, 53, 255)


def fit(image: Image.Image, box: tuple[int, int]) -> Image.Image:
    scale = min(box[0] / image.width, box[1] / image.height)
    size = (max(1, round(image.width * scale)), max(1, round(image.height * scale)))
    return image.resize(size, Image.Resampling.LANCZOS)


def centred(size: int, logo: Image.Image, fraction: float) -> Image.Image:
    canvas = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    mark = fit(logo, (round(size * fraction), round(size * fraction)))
    canvas.alpha_composite(mark, ((size - mark.width) // 2, (size - mark.height) // 2))
    return canvas


def main() -> None:
    if not LOGO.exists():
        raise SystemExit(f"missing J3 brand logo: {LOGO}")
    logo = Image.open(LOGO).convert("RGBA")

    for density, legacy, foreground in [
        ("mdpi", 48, 108),
        ("hdpi", 72, 162),
        ("xhdpi", 96, 216),
        ("xxhdpi", 144, 324),
        ("xxxhdpi", 192, 432),
    ]:
        folder = RES / f"mipmap-{density}"

        square = Image.new("RGBA", (legacy, legacy), NAVY)
        square.alpha_composite(centred(legacy, logo, 0.72))
        square.convert("RGB").save(folder / "ic_launcher.png", optimize=True)

        mask = Image.new("L", (legacy, legacy), 0)
        ImageDraw.Draw(mask).ellipse((0, 0, legacy - 1, legacy - 1), fill=255)
        round_icon = Image.new("RGBA", (legacy, legacy), NAVY)
        round_icon.alpha_composite(centred(legacy, logo, 0.72))
        round_icon.putalpha(mask)
        round_icon.save(folder / "ic_launcher_round.png", optimize=True)

        centred(foreground, logo, 0.62).save(
            folder / "ic_launcher_foreground.png", optimize=True
        )

    for folder, size in [
        ("drawable", 480),
        ("drawable-port-mdpi", 480),
        ("drawable-land-mdpi", 480),
        ("drawable-port-hdpi", 720),
        ("drawable-land-hdpi", 720),
        ("drawable-port-xhdpi", 960),
        ("drawable-land-xhdpi", 960),
        ("drawable-port-xxhdpi", 1440),
        ("drawable-land-xxhdpi", 1440),
        ("drawable-port-xxxhdpi", 1920),
        ("drawable-land-xxxhdpi", 1920),
    ]:
        path = RES / folder / "splash.png"
        if not path.exists():
            continue
        width, height = size, size
        if "land" in folder:
            width = int(size * 1.6)
        canvas = Image.new("RGB", (width, height), NAVY[:3])
        mark = fit(logo, (round(min(width, height) * 0.42), round(min(width, height) * 0.42)))
        canvas.paste(mark, ((width - mark.width) // 2, (height - mark.height) // 2), mark)
        canvas.save(path, optimize=True)

    print("J3 launcher and splash assets generated")


if __name__ == "__main__":
    main()
