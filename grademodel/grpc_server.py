"""
gRPC server for Complexity Grader
"""

import grpc
from concurrent import futures
import grade_pb2
import grade_pb2_grpc
from grader import ComplexityGrader


class GraderServicer(grade_pb2_grpc.GraderServicer):
    def __init__(self):
        self.grader = ComplexityGrader()
        self.grader.load("model.pkl")
        print("✅ ComplexityGrader model loaded")

    def GradeTask(self, request, context):
        grade = self.grader.predict(request.task)
        translated = self.grader._translate(request.task)
        return grade_pb2.GradeResponse(
            task=request.task,
            translated=translated,
            grade=grade
        )


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    grade_pb2_grpc.add_GraderServicer_to_server(GraderServicer(), server)
    server.add_insecure_port('[::]:50054')
    print("🚀 Grader gRPC server started on port 50054")
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
