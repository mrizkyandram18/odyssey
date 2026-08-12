import re
with open('pkg/game/quest/handler_test.go', 'r') as f:
    text = f.read()
text = re.sub(r'(h\.CompleteChallenge\(context\.Background\(\), 1, 12, "crew-1", "user-1", ""\))', r'_, err := \1\n\t\t\tif err != nil {\n\t\t\t\tt.Logf("err: %v", err)\n\t\t\t}', text)
with open('pkg/game/quest/handler_test.go', 'w') as f:
    f.write(text)
