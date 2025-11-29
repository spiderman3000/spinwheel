import React, { useState } from 'react';

interface AddItemFormProps {
    addItem: (option: string) => void;
}

export const AddItemForm: React.FC<AddItemFormProps> = ({ addItem }) => {
    const [text, setText] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (text.trim()) {
            addItem(text.trim());
            setText('');
        }
    };

    return (
        <form onSubmit={handleSubmit} className="w-full max-w-sm md:max-w-md flex gap-2">
            <input
                type="text"
                value={text}
                onChange={(e) => setText(e.target.value)}
                placeholder="Add a new option"
                className="w-full px-4 py-2 rounded-lg bg-fill text-text-primary placeholder-text-secondary border border-border focus:outline-none focus:ring-2 focus:ring-accent transition-colors"
            />
            <button
                type="submit"
                className="px-6 py-2 rounded-lg bg-accent text-white font-semibold shadow-md hover:opacity-90 transition-opacity active:scale-95"
            >
                Add
            </button>
        </form>
    );
};