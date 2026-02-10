import React, { createContext, useContext, useCallback } from "react";

type NavigateFunction = (pageId: string) => void;

const NavigationContext = createContext<NavigateFunction | null>(null);

export const NavigationProvider: React.FC<{
  children: React.ReactNode;
  onNavigate: NavigateFunction;
}> = ({ children, onNavigate }) => {
  return (
    <NavigationContext.Provider value={onNavigate}>
      {children}
    </NavigationContext.Provider>
  );
};

export const useNavigation = () => {
  const context = useContext(NavigationContext);
  if (!context) {
    throw new Error("useNavigation must be used within NavigationProvider");
  }
  return context;
};

export const useNavigateHome = () => {
  const navigate = useNavigation();
  return useCallback(() => navigate("home"), [navigate]);
};
