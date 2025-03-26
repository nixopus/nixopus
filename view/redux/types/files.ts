export interface FileData {
  path: string;
  name: string;
  size: number;
  created_at: string;
  updated_at: string;
  file_type: FileType;
  permissions: number;
  is_hidden: boolean;
  extension: string | null;
  owner_id: number;
  group_id: number;
}

export enum FileType {
  File = 'File',
  Directory = 'Directory',
  Symlink = 'Symlink',
  Other = 'Other'
}
