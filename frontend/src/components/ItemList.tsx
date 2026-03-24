import type { Item } from '../App';
import { ItemRow } from './Item';

interface ItemListProps {
    items: Item[];
    deleteItem: (id: number) => void;
}

export const ItemList = ({ items, deleteItem }: ItemListProps) => {
    return (
        <div className="card">
            <h2 style={{ marginBottom: '1rem', fontSize: '1.25rem' }}>Options</h2>
            <div className="item-list">
                {items.length === 0 ? (
                    <p style={{ color: 'var(--text-secondary)', textAlign: 'center' }}>No options added yet.</p>
                ) : (
                    items.map((item) => (
                        <ItemRow key={item.id} item={item} deleteItem={deleteItem} />
                    ))
                )}
            </div>
        </div>
    );
};
