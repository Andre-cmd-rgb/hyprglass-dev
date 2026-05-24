#!/usr/bin/env python3
from PIL import Image, ImageDraw, ImageFilter
from pathlib import Path
import math, random
W,H=3840,2160
img=Image.new('RGB',(W,H))
p=img.load()
for y in range(H):
    t=y/H
    r=int(7+(42-7)*t); g=int(11+(18-11)*t); b=int(35+(62-35)*t)
    if y < H*0.42:
        b += int((1-t)*35); r += int((1-t)*22)
    for x in range(W): p[x,y]=(r,g,b)
d=ImageDraw.Draw(img,'RGBA')
random.seed(23)
for _ in range(220):
    x=random.randrange(W); y=random.randrange(int(H*.05), int(H*.48)); a=random.randrange(55,150); s=random.choice([1,1,2])
    d.ellipse((x,y,x+s,y+s),fill=(219,234,254,a))
# moon
d.ellipse((2920,250,3070,400),fill=(224,231,255,230))
d.ellipse((2960,225,3110,390),fill=(23,18,54,255))
# mountains
for layer,(base,col) in enumerate([(1240,(24,30,62,230)),(1390,(13,18,38,245)),(1540,(7,11,26,255))]):
    pts=[(0,H)]
    for x in range(0,W+240,240):
        y=base-int(math.sin(x/310+layer)*105)-random.randrange(25,180)
        pts.append((x,y))
    pts += [(W,H)]
    d.polygon(pts,fill=col)
# lake
d.rectangle((0,1540,W,1900),fill=(10,18,42,185))
for i in range(90):
    y=1560+i*4+random.randrange(0,4); x=random.randrange(0,W-300); l=random.randrange(80,520)
    d.line((x,y,x+l,y),fill=(96,165,250,random.randrange(18,65)),width=random.choice([1,1,2]))
for _ in range(120):
    x=random.randrange(300,W-300); y=random.randrange(1340,1680)
    d.ellipse((x,y,x+3,y+3),fill=(250,204,21,random.randrange(80,180)))
# foreground
for x in range(0,W,55):
    h=random.randrange(90,310); base=H
    d.polygon([(x,base),(x+24,base-h),(x+48,base)],fill=(3,7,18,255))
    d.rectangle((x+22,base-h,x+27,base),fill=(3,7,18,255))
img=img.filter(ImageFilter.GaussianBlur(0.25))
out=Path(__file__).resolve().parents[1]/'assets/wallpapers/hyprglass-dusk.png'
out.parent.mkdir(parents=True,exist_ok=True)
img.save(out)
print(out)
