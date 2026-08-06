import io

with io.open('fake_key.txt', 'r', encoding='utf-16le') as f:
    key = f.read().strip()

with io.open('.env', 'r', encoding='utf-8') as f:
    env = f.read()

env = env.replace('"project_id":"dummy"', f'"project_id":"dummy","private_key":"{key}"')

with io.open('.env', 'w', encoding='utf-8') as f:
    f.write(env)
