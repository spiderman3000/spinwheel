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
            <h1 id="playground-heading" className="playground-heading">
                Random spin wheel — add options and spin to pick one
            </h1>
            <div className="playground-body">
                <div className="sidebar">
                    <AddItemForm addItem={addItem} />
                    <ItemList items={items} deleteItem={deleteItem} />
                </div>
                <Wheel items={items} isDarkMode={isDarkMode} />
            </div>
        </div>
    );
};

export default Playground;
