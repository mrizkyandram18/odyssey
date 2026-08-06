import io
import re

with io.open(r'D:\Personal\Projects\kuota - Copy\.env', 'r', encoding='utf-8') as f:
    kuota = f.read()

fb_cred = re.search(r'FIREBASE_CREDENTIALS=".*?"', kuota, re.DOTALL).group(0)

with io.open('.env', 'r', encoding='utf-8') as f:
    env = f.read()

env = re.sub(r'FIREBASE_CREDENTIALS=.*', fb_cred, env)

with io.open('.env', 'w', encoding='utf-8') as f:
    f.write(env)
