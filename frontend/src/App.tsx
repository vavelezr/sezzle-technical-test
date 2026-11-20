import styles from './App.module.css';
import { Calculator } from './features/calculator/components/Calculator';

export function App() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title}>Calculator</h1>
        <p className={styles.caption}>Built for Sezzle technical assessment</p>
      </header>
      <main className={styles.main}>
        <Calculator />
      </main>
    </div>
  );
}
