import {
  Folder,
  FileCode,
  FileSpreadsheet,
  FileAudio,
  FileVideo,
  FileImage,
  FileText,
  FileArchive,
  FileCheck,
  FileX,
  FileCog,
  FileJson,
  FileType,
  Database,
  BrainCircuit,
  Settings
} from 'lucide-react';

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

type IconProps = React.ComponentProps<typeof FileCode>;

export function getFileIcons(
  type: 'folder' | 'file',
  ext: string,
  layout: 'grid' | 'list' | 'gallery'
): React.ReactNode {
  const iconSize = layout === 'grid' ? 'h-10 w-10 lg:h-20 lg:w-20' : 'h-5 w-5 lg:h-8 lg:w-8';
  const baseProps: IconProps = {
    className: `${iconSize}`
  };

  if (type === 'folder') {
    return <Folder {...baseProps} className={`${iconSize} fill-[#58b7e9] text-[#58b7e9] `} />;
  } else {
    switch (ext.toLowerCase()) {
      case 'js':
      case 'ts':
      case 'jsx':
      case 'tsx':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'py':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'java':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'rb':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'php':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'cs':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'go':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'rs':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'html':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'css':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'scss':
      case 'sass':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'json':
        return <FileJson {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'yaml':
      case 'yml':
        return <FileType {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'xml':
        return <FileCode {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'csv':
      case 'tsv':
      case 'xlsx':
      case 'xls':
        return <FileSpreadsheet {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'mp3':
      case 'wav':
      case 'ogg':
      case 'flac':
      case 'aac':
        return <FileAudio {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'mp4':
      case 'avi':
      case 'mov':
      case 'mkv':
      case 'webm':
        return <FileVideo {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'bmp':
      case 'svg':
      case 'webp':
        return <FileImage {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'pdf':
        return <FileText {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'doc':
      case 'docx':
        return <FileText {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'txt':
      case 'md':
        return <FileText {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'zip':
      case 'rar':
      case '7z':
      case 'tar':
      case 'gz':
        return <FileArchive {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'exe':
      case 'app':
      case 'dmg':
        return <FileCog {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'sql':
      case 'db':
      case 'sqlite':
        return <Database {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'pkl':
      case 'h5':
      case 'onnx':
        return <BrainCircuit {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'config':
      case 'ini':
      case 'env':
        return <Settings {...baseProps} className={`${iconSize} text-secondary`} />;

      case 'gitignore':
      case 'npmignore':
        return <FileX {...baseProps} className={`${iconSize} text-secondary`} />;
      case 'license':
      case 'dockerfile':
        return <FileCheck {...baseProps} className={`${iconSize} text-secondary`} />;

      default:
        return <FileText {...baseProps} className={`${iconSize} text-secondary`} />;
    }
  }
}

export const parsePort = (port: string) => {
  const parsedPort = parseInt(port, 10);
  return isNaN(parsedPort) ? null : parsedPort;
};

export const SHARED_DOMAIN_REGEX = /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;
