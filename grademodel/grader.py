"""
Lightweight model for grading task complexity (1-100)
Uses Google Translate + keyword features + Ridge regression
"""

import asyncio
import pickle
import numpy as np
from googletrans import Translator, constants
from sklearn.linear_model import Ridge


class ComplexityGrader:
    def __init__(self):
        self.model = Ridge(alpha=1.0)
        self.translator = Translator()
        self._is_trained = False

    def _translate(self, text, src='auto'):
        try:
            result = asyncio.run(self.translator.translate(text, dest='en'))
            return result.text.lower() if result else text.lower()
        except Exception:
            return text.lower()

    def _extract_features(self, text):
        f = {}
        text_lower = text.lower()
        words = text_lower.split()
        
        f['word_count'] = len(words)
        f['char_count'] = len(text)
        f['avg_word_len'] = f['char_count'] / max(f['word_count'], 1)
        
        keywords = {
            'has_code': ['func', 'class', 'def ', 'function', 'method', 'script'],
            'has_api': ['api', 'http', 'request', 'endpoint', 'rest', 'graphql', 'microservice'],
            'has_db': ['sql', 'query', 'database', 'postgresql', 'mysql', 'redis', 'mongo'],
            'has_auth': ['auth', 'login', 'password', 'token', 'jwt', 'oauth'],
            'has_async': ['async', 'await', 'websocket', 'streaming', 'real-time'],
            'has_test': ['test', 'mock', 'unittest'],
            'has_frontend': ['react', 'vue', 'angular', 'css', 'html', 'ui', 'frontend', 'site', 'page'],
            'has_devops': ['docker', 'kubernetes', 'k8s', 'deploy', 'ci/cd', 'nginx', 'heroku'],
            'has_ml': ['machine learning', 'ml', 'tensorflow', 'pytorch', 'neural'],
            'has_payment': ['payment', 'stripe', 'paypal'],
            'has_video': ['video', 'stream', 'webrtc'],
            'has_blockchain': ['blockchain'],
            'has_iot': ['IoT', 'iot', 'sensor'],
            'has_go': ['golang', 'go '],
            'has_proxy': ['proxy'],
            'has_mobile': ['ios', 'android', 'flutter', 'react native'],
            'is_simple': ['simple', 'mini', 'hello world', 'calculator'],
            'is_medium': ['site', 'app', 'dashboard', 'blog'],
            'is_complex': ['system', 'platform', 'marketplace'],
            'no_frontend': ['without frontend', 'no frontend', 'without ui'],
            'no_test': ['without test', 'no test'],
        }
        
        for name, kw_list in keywords.items():
            f[name] = 1 if any(k in text_lower for k in kw_list) else 0
        
        return f

    def train(self, tasks_texts, grades, epochs=100):
        if len(tasks_texts) != len(grades):
            raise ValueError("Tasks and grades must have same length")
        
        print("Translating tasks to English...")
        en_texts = []
        for i, t in enumerate(tasks_texts):
            en = self._translate(t)
            en_texts.append(en)
            if (i + 1) % 10 == 0:
                print(f"  Translated {i + 1}/{len(tasks_texts)}")
        
        extra_features = [list(self._extract_features(t).values()) for t in en_texts]
        X = np.array(extra_features)
        y = np.array(grades)
        
        for epoch in range(epochs):
            indices = np.random.permutation(len(X))
            self.model = Ridge(alpha=1.0)
            self.model.fit(X[indices], y[indices])
            if epochs <= 50 or (epoch + 1) % 10 == 0:
                print(f"  Epoch {epoch + 1}/{epochs}")
        
        self._is_trained = True
        print(f"Model trained on {len(tasks_texts)} samples, {epochs} epochs")

    def predict(self, task_text):
        if not self._is_trained:
            raise ValueError("Model not trained yet")
        
        en = self._translate(task_text)
        feat = self._extract_features(en)
        X = np.array([list(feat.values())])
        
        pred = self.model.predict(X)[0]
        return max(1, min(100, round(pred)))

    def save(self, path):
        with open(path, 'wb') as f:
            pickle.dump({
                'model': self.model,
                'is_trained': self._is_trained
            }, f)
        print(f"Model saved to {path}")

    def load(self, path):
        with open(path, 'rb') as f:
            data = pickle.load(f)
        self.model = data['model']
        self._is_trained = data['is_trained']
        print(f"Model loaded from {path}")