import React from 'react';
import type { Item } from '../App';
import { ItemRow } from './Item';

interface ItemListProps {
    items: Item[];
    deleteItem: (id: number) => void;
}

export const ItemList: React.FC<ItemListProps> = ({ items, deleteItem }) => {
    return (
        <div className="w-full max-w-sm md:max-w-md bg-fill border border-border rounded-2xl p-4 transition-colors">
            <h2 className="text-xl font-semibold text-text-primary mb-4">Options</h2>
            <div className="flex flex-col gap-2 max-h-60 overflow-y-auto">
                {items.map((item) => (
                    <ItemRow key={item.id} item={item} deleteItem={deleteItem} />
                ))}
            </div>
        </div>
    );
};