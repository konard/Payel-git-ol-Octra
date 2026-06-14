import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import type { GetProjectFilesResponse } from "./user-files_pb";

const transport = createConnectTransport({
  baseUrl: (import.meta as any).env?.VITE_GRPC_WEB_URL || "http://localhost:50052",
});

interface ServiceDefinition {
  typeName: string;
  methods: Record<string, {
    name: string;
    I: { typeName: string };
    O: { typeName: string };
    kind: "unary";
  }>;
}

const UserFilesService: ServiceDefinition = {
  typeName: "userfiles.UserFilesService",
  methods: {
    getProjectFiles: {
      name: "GetProjectFiles",
      I: { typeName: "userfiles.GetProjectFilesRequest" },
      O: { typeName: "userfiles.GetProjectFilesResponse" },
      kind: "unary",
    },
  },
};

export const userFilesClient = createClient(UserFilesService as any, transport);

export async function getProjectFiles(chatId: string): Promise<GetProjectFilesResponse> {
  const response = await (userFilesClient as any).getProjectFiles({ chatId });
  return response as GetProjectFilesResponse;
}
