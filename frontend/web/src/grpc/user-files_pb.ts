export interface GetProjectFilesRequest {
  chatId: string;
}

export interface CodeFileEntry {
  path: string;
  content: string;
  language: string;
  encoding: string;
}

export interface GetProjectFilesResponse {
  chatId: string;
  taskId: string;
  files: CodeFileEntry[];
  totalFiles: number;
}
