import { useState, useRef, DragEvent, ChangeEvent } from 'react';

interface FileObject {
  file: File;
  id: string;
  progress: number;
  uploading: boolean;
}
function use_file_upload() {
  const [files, setFiles] = useState<FileObject[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragEnter = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    handleFiles(droppedFiles);
  };

  const handleFileInput = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files);
      handleFiles(selectedFiles);
    }
  };

  const handleFiles = (newFiles: File[]) => {
    const updatedFiles: FileObject[] = newFiles.map((file) => ({
      file,
      id: Math.random().toString(36).substr(2, 9),
      progress: 0,
      uploading: false
    }));
    setFiles((prev) => [...prev, ...updatedFiles]);
    updatedFiles.forEach((fileObj) => simulateUpload(fileObj.id));
  };

  const simulateUpload = (id: string) => {
    let progress = 0;
    const interval = setInterval(() => {
      progress += 10;
      setFiles((prev) => prev.map((f) => (f.id === id ? { ...f, progress, uploading: true } : f)));
      if (progress >= 100) {
        clearInterval(interval);
        setFiles((prev) => prev.map((f) => (f.id === id ? { ...f, uploading: false } : f)));
      }
    }, 500);
  };

  const deleteFile = (id: string) => {
    setFiles((prev) => prev.filter((f) => f.id !== id));
  };

  return {
    files,
    isDragging,
    fileInputRef,
    handleDragEnter,
    handleDragLeave,
    handleDragOver,
    handleDrop,
    handleFileInput,
    handleFiles,
    simulateUpload,
    deleteFile
  };
}

export default use_file_upload;
