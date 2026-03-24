import { useState, useEffect } from 'preact/hooks';
import './App.css';
import { Navbar, Playground } from './components/index';

export interface Item {
  id: number;
  option: string;
}

const initialItems: Item[] = [
  { id: 1, option: 'Preact' },
  { id: 2, option: 'Vite' },
  { id: 3, option: 'Fast' },
  { id: 4, option: 'Minimal' },
];

function App() {
  const [items, setItems] = useState<Item[]>(initialItems);
  const [isDarkMode, setIsDarkMode] = useState(false);

  useEffect(() => {
    const isDark = localStorage.getItem('theme') === 'dark';
    setIsDarkMode(isDark);
    if (isDark) {
        document.documentElement.classList.add('dark');
    } else {
        document.documentElement.classList.remove('dark');
    }
  }, []);

  const toggleTheme = () => {
      const newIsDarkMode = !isDarkMode;
      setIsDarkMode(newIsDarkMode);
      if (newIsDarkMode) {
          document.documentElement.classList.add('dark');
          localStorage.setItem('theme', 'dark');
      } else {
          document.documentElement.classList.remove('dark');
          localStorage.setItem('theme', 'light');
      }
  };

  const addItem = (option: string) => {
    const newItem: Item = {
      id: Date.now(),
      option,
    };
    setItems((prevItems) => [...prevItems, newItem]);
  };

  const deleteItem = (id: number) => {
    setItems((prevItems) => prevItems.filter((item) => item.id !== id));
  };

  return (
    <div className="app-container">
      <Navbar isDarkMode={isDarkMode} toggleTheme={toggleTheme} />
      <main className="main-content">
        <Playground items={items} addItem={addItem} deleteItem={deleteItem} isDarkMode={isDarkMode} />
      </main>
    </div>
  );
}

export default App;
