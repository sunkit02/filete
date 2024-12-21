import { cn, getPreferedTheme, setPreferedTheme } from "@/utils";
import { useEffect, useState } from "react";
import { newAuthContext } from "./lib/services/auth";
import UploadPage from "./components/UploadPage";
import { ThemeProvider } from "./providers/theme-provider";
import { ThemeToggle } from "./components/ThemeToggle";

function App() {
  useEffect(() => {
    setPreferedTheme(getPreferedTheme());
  }, []);

  const [authContext, setAuthContext] = useState(newAuthContext());

  return (
    <ThemeProvider defaultTheme="dark" storageKey="theme">
      <main
        className={cn(
          "h-[100vh] w-[100vw] flex flex-col items-center justify-start p-5 gap-4 bg-white dark:bg-black",
        )}
      >
        <h1 className="dark:text-white font-bold text-5xl">
          Hello, {authContext.username}!
        </h1>
        {/*
        <button
          className={cn("px-5 py-2 rounded bg-blue-700 text-white", {
            "bg-blue-300": !isLight,
          })}
          onClick={handleToggleDarkMode}
          type="button"
        >
          {isLight ? "Dark" : "Light"}
        </button>
        */}
        <ThemeToggle />
        <UploadPage />
      </main>
    </ThemeProvider>
  );
}

export default App;
