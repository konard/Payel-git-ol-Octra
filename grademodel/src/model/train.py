"""
Train multilingual grader on tasks.txt data
"""

import argparse
from grademodel.src.model.grader import ComplexityGrader

def parse_tasks(filepath):
    tasks = []
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    for i in range(0, len(lines) - 1, 2):
        task_text = lines[i].strip()
        task_text = task_text.split('. ', 1)[1] if '. ' in task_text else task_text
        grade_line = lines[i + 1].strip()
        
        if grade_line.startswith('grade:'):
            grade = int(grade_line.split(':')[1].strip())
            tasks.append((task_text, grade))
    
    return tasks

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--epochs', type=int, default=100, help='Number of training epochs')
    args = parser.parse_args()
    
    tasks = parse_tasks('../../tasks.txt')
    print(f"Loaded {len(tasks)} tasks")
    
    texts = [t[0] for t in tasks]
    grades = [t[1] for t in tasks]
    
    grader = ComplexityGrader()
    grader.train(texts, grades, epochs=args.epochs)
    
    print("\n--- Predictions ---")
    tasks_test = [
        ("Создать простую функцию Hello World", "create simple function"),
        ("Настроить Kubernetes кластер", "set up Kubernetes cluster"),
        ("Реализовать нейронную сеть с нуля", "implement neural network from scratch"),
        ("Нужен мини прокси на go без фронтенда", "mini proxy in go without frontend"),
        ("需要一个用Go编写的迷你代理，不需要前端", "mini proxy in Go, no frontend needed"),
    ]
    
    for ru, en in tasks_test:
        g_ru = grader.predict(ru)
        g_en = grader.predict(en)
        print(f"[{g_ru:>2}/100] [{g_en:>2}/100] {en[:30]}")
    
    grader.save("model.pkl")