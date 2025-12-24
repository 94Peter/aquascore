import React, { createContext, useState, useContext, ReactNode } from 'react';

interface AthleteContextType {
  selectedAthlete: string | null;
  setSelectedAthlete: (athlete: string | null) => void;
}

const AthleteContext = createContext<AthleteContextType | undefined>(undefined);

export const AthleteProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [selectedAthlete, setSelectedAthlete] = useState<string | null>(null); // Default to null

  return (
    <AthleteContext.Provider value={{ selectedAthlete, setSelectedAthlete }}>
      {children}
    </AthleteContext.Provider>
  );
};

// eslint-disable-next-line react-refresh/only-export-components
export const useAthlete = () => {
  const context = useContext(AthleteContext);
  if (context === undefined) {
    throw new Error('useAthlete must be used within an AthleteProvider');
  }
  return context;
};