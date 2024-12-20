import { cn, getPreferedTheme, setPreferedTheme } from "@/utils";
import { useEffect, useState } from "react";

function App() {
  useEffect(() => {
    setPreferedTheme(getPreferedTheme());
  }, []);

  const [isLight, setIsLight] = useState<boolean>(
    getPreferedTheme() === "light",
  );

  const handleToggleDarkMode = () => {
    const newTheme = getPreferedTheme() === "light" ? "dark" : "light";
    setPreferedTheme(newTheme);

    setIsLight(newTheme === "light");
  };

  return (
    <main
      className={cn(
        "h-[100vh] w-[100vw] flex flex-col items-center justify-start p-5 gap-4 bg-white dark:bg-black",
      )}
    >
      <h1 className="text-blue-700 dark:text-white font-bold text-5xl">
        Hello, Filete!
      </h1>
      <button
        className={cn("px-5 py-2 rounded bg-blue-700 text-white", {
          "bg-blue-300": !isLight,
        })}
        onClick={handleToggleDarkMode}
        type="button"
      >
        {isLight ? "Dark" : "Light"}
      </button>
    </main>
  );
}

export default App;
