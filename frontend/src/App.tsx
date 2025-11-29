import { useState, useEffect } from 'react';
import './App.css';
import { Navbar, Playground } from './components/index';

export interface Item {
  id: number;
  option: string;
}

const initialItems: Item[] = [
  { id: 1, option: 'React' },
  { id: 2, option: 'Vue' },
  { id: 3, option: 'Angular' },
  { id: 4, option: 'Svelte' },
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
    <div className="bg-background text-text-primary min-h-screen">
      <Navbar isDarkMode={isDarkMode} toggleTheme={toggleTheme} />
      <main className="max-w-7xl mx-auto p-4 md:p-8">
        <Playground items={items} addItem={addItem} deleteItem={deleteItem} isDarkMode={isDarkMode} />
      </main>
    </div>
  );
}

export default App;
