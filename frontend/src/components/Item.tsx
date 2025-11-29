import React from "react";
import type { Item } from "../App";

interface ItemRowProps {
    item: Item;
    deleteItem: (id: number) => void;
}

export const ItemRow: React.FC<ItemRowProps> = ({ item, deleteItem }) => {
    return (
        <div className="group flex items-center justify-between border-b border-border p-3 transition-colors">
            <span className="text-text-secondary">{item.option}</span>
            <button
                onClick={() => deleteItem(item.id)}
                className="p-1 rounded-full text-text-secondary opacity-0 group-hover:opacity-100 hover:text-text-primary transition-opacity"
            >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
            </button>
        </div>
    );
};