"""Test grader translations"""
from grader import ComplexityGrader

g = ComplexityGrader()
g.load("model.pkl")

tests = [
    "Нужен мини прокси на go без фронтенда",
    "需要一个用Go编写的迷你代理，不需要前端",
    "需要一个用 Go 语言编写的、无需前端或测试的迷你代理",
]

print("\n--- Results ---")
for t in tests:
    en = g._translate(t)
    grade = g.predict(t)
    print(f"EN: {en}")
    print(f"Grade: {grade}/100\n")