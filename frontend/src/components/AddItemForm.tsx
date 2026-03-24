import { useState } from 'preact/hooks';

interface AddItemFormProps {
    addItem: (option: string) => void;
}

const AddItemForm = ({ addItem }: AddItemFormProps) => {
    const [inputValue, setInputValue] = useState('');

    const handleSubmit = (e: Event) => {
        e.preventDefault();
        if (inputValue.trim()) {
            addItem(inputValue.trim());
            setInputValue('');
        }
    };

    return (
        <form onSubmit={handleSubmit} className="card">
            <h2 style={{ marginBottom: '1rem', fontSize: '1.25rem' }}>Add Option</h2>
            <div className="input-group">
                <input
                    type="text"
                    value={inputValue}
                    onInput={(e) => setInputValue((e.target as HTMLInputElement).value)}
                    placeholder="Enter an option..."
                />
                <button type="submit" className="btn btn-primary">
                    Add
                </button>
            </div>
        </form>
    );
};

export default AddItemForm;
