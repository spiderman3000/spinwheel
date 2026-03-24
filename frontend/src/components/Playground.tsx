import type { Item } from '../App';
import AddItemForm from './AddItemForm';
import { ItemList } from './ItemList';
import Wheel from './Wheel';

interface PlaygroundProps {
    items: Item[];
    addItem: (option: string) => void;
    deleteItem: (id: number) => void;
    isDarkMode: boolean;
}

const Playground = ({ items, addItem, deleteItem, isDarkMode }: PlaygroundProps) => {
    return (
        <div className="playground">
            <div className="sidebar">
                <AddItemForm addItem={addItem} />
                <ItemList items={items} deleteItem={deleteItem} />
            </div>
            <Wheel items={items} isDarkMode={isDarkMode} />
        </div>
    );
};

export default Playground;
