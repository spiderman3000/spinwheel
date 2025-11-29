import type { Item } from '../App';
import { AddItemForm } from './AddItemForm';
import { ItemList } from './ItemList';
import WheelComponent from './Wheel';

interface PlaygroundProps {
    items: Item[];
    addItem: (option: string) => void;
    deleteItem: (id: number) => void;
    isDarkMode: boolean;
}

export const Playground: React.FC<PlaygroundProps> = ({ items, addItem, deleteItem, isDarkMode }) => {
    return (
        <div className="w-full inset-0 flex flex-col md:flex-row justify-center items-center gap-8 overflow-hidden">
            <WheelComponent items={items} isDarkMode={isDarkMode} />
            <div className="w-full md:w-auto flex flex-col gap-4">
                <AddItemForm addItem={addItem} />
                <ItemList items={items} deleteItem={deleteItem} />
            </div>
        </div>
    );
};
