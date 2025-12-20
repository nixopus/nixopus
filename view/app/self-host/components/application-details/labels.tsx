'use client';
import React, { useEffect, useRef, useState } from 'react';
import { Input } from '@/components/ui/input';
import { Plus } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { useUpdateApplicationLabelsMutation } from '@/redux/services/deploy/applicationsApi';

interface LabelsProps {
	applicationId: string;
	labels?: string[];
	onLabelsChange?: (labels: string[]) => void;
	isEditable?: boolean;
}

export default function Labels({ applicationId, labels = [], onLabelsChange, isEditable = true }: LabelsProps) {
	const [localLabels, setLocalLabels] = useState<string[]>(labels);
	const [isAdding, setIsAdding] = useState(false);
	const [newLabel, setNewLabel] = useState('');
	const inputRef = useRef<HTMLInputElement>(null);
	const [updateLabels, { isLoading }] = useUpdateApplicationLabelsMutation();

	useEffect(() => {
		setLocalLabels(labels);
	}, [labels]);

	useEffect(() => {
		if (isAdding && inputRef.current) inputRef.current.focus();
	}, [isAdding]);

	const saveLabels = async (updated: string[]) => {
		await updateLabels({ id: applicationId, labels: updated }).unwrap();
		setLocalLabels(updated);
		onLabelsChange?.(updated);
	};

	const handleAdd = async () => {
		const value = newLabel.trim();
		if (!value) {
			setIsAdding(false);
			setNewLabel('');
			return;
		}
		const updated = [...new Set([...localLabels, value])];
		await saveLabels(updated);
		setIsAdding(false);
		setNewLabel('');
	};

	const handleKeyDown = async (e: React.KeyboardEvent<HTMLInputElement>) => {
		if (e.key === 'Enter') {
			await handleAdd();
		} else if (e.key === 'Escape') {
			setIsAdding(false);
			setNewLabel('');
		}
	};

	return (
		<div className="flex flex-wrap gap-2">
			{isEditable ? (
				<>
					{isAdding ? (
						<Input
							ref={inputRef}
							value={newLabel}
							onChange={(e) => setNewLabel(e.target.value)}
							onKeyDown={handleKeyDown}
							onBlur={handleAdd}
							className="h-6 w-28 text-xs"
							placeholder="New label"
							disabled={isLoading}
						/>
					) : (
						<button
							type="button"
							onClick={() => setIsAdding(true)}
							className="inline-flex items-center gap-1 h-5 px-2 rounded-md border text-xs text-muted-foreground hover:bg-secondary"
						>
							<Plus size={12} />
							Add Label
						</button>
					)}
				</>
			) : (
				<>
					{localLabels.map((label, index) => (
						<Badge key={`${label}-${index}`} variant="secondary" className="text-xs px-2 py-1 h-5">
							{label}
						</Badge>
					))}
				</>
			)}
		</div>
	);
}
