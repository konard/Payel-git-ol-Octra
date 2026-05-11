"""
FastAPI server for complexity grader
"""

from fastapi import FastAPI
from pydantic import BaseModel
from grader import ComplexityGrader

app = FastAPI(title="Complexity Grader API")

grader = ComplexityGrader()
grader.load("model.pkl")


class TaskRequest(BaseModel):
    task: str


class GradeResponse(BaseModel):
    task: str
    translated: str
    grade: int


@app.get("/")
def root():
    return {"message": "Complexity Grader API"}


@app.post("/grade", response_model=GradeResponse)
def grade_task(request: TaskRequest):
    translated = grader._translate(request.task)
    grade = grader.predict(request.task)
    return GradeResponse(task=request.task, translated=translated, grade=grade)