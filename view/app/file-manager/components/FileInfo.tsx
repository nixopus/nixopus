import React from 'react';
import { DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { FileData } from '@/redux/types/files';
import { formatFileSize } from '@/app/self-host/utils/formatFileSize';

interface FileInfoProps {
    file: FileData;
    isLoading: boolean;
    fileSize: number | null;
}

function FileInfo({ file, isLoading, fileSize }: FileInfoProps) {
    return (
        <DialogContent className="sm:max-w-[425px]">
            <DialogHeader className="space-y-6">
                <DialogTitle className="text-2xl font-bold text-primary">{file.name}</DialogTitle>
                <div className="mx-auto w-full space-y-4 rounded-lg bg-secondary/10 p-4">
                    <InfoItem label="Path" value={file.path} />
                    <InfoItem
                        label="Size"
                        value={isLoading ? 'Calculating...' : formatFileSize(fileSize || file.size)}
                    />
                    <InfoItem label="Created" value={new Date(file.created_at).toLocaleString()} />
                    <InfoItem label="Updated" value={new Date(file.updated_at).toLocaleString()} />
                    <InfoItem label="Type" value={file.file_type} />
                    <InfoItem label="Hidden" value={file.is_hidden ? 'Yes' : 'No'} />
                    <InfoItem
                        label="Permissions"
                        value={
                            (file.permissions & 0o400 ? 'r' : '-') +
                            (file.permissions & 0o200 ? 'w' : '-') +
                            (file.permissions & 0o100 ? 'x' : '-')
                        }
                    />
                </div>
            </DialogHeader>
        </DialogContent>
    );
}

function InfoItem({ label, value }: { label: string; value: string }) {
    return (
        <div className="flex items-center space-x-3 text-sm">
            <span className="font-medium text-secondary-foreground">{label}:</span>
            <span className="text-secondary-foreground/80">{value}</span>
        </div>
    );
}

export default FileInfo;
