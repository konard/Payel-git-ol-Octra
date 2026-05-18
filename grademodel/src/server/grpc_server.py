"""
gRPC server for Complexity Grader
"""

from concurrent import futures
import grpc

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    grade_pb2_grpc.add_GraderServicer_to_server(GraderServicer(), server)
    server.add_insecure_port('[::]:50054')
    print("🚀 Grader gRPC server started on port 50054")
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
