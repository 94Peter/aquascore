import React, { useState } from 'react';
import { Search, Check, ChevronsUpDown } from 'lucide-react';
import { Combobox } from '@headlessui/react';
import clsx from 'clsx';

interface AthleteSelectorProps {
  onSelectAthlete: (athleteName: string | null) => void;
  selectedAthlete: string | null;
  athletes: readonly string[] | undefined;
  isLoading: boolean;
}

const DISPLAY_LIMIT = 50; // Limit the number of displayed athletes

const AthleteSelector: React.FC<AthleteSelectorProps> = ({
  onSelectAthlete,
  selectedAthlete,
  athletes,
  isLoading,
}) => {
  const [query, setQuery] = useState('');

  const handleSelection = (athlete: string | null) => {
    console.log('SELECTION CHANGE:', athlete);
    onSelectAthlete(athlete);
    // Reset query after selection
    setQuery('');
  };
  
  const handleQueryChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    console.log('QUERY CHANGE:', event.target.value);
    setQuery(event.target.value);
  };

  const filteredAthletes =
    query === ''
      ? athletes // If query is empty, show all athletes
      : athletes?.filter((athlete) => {
          return athlete.toLowerCase().includes(query.toLowerCase());
        });

  const displayedAthletes = filteredAthletes?.slice(0, DISPLAY_LIMIT);

  if (isLoading) {
    return (
      <div>
        <label className="text-sm font-medium text-slate-600">運動員姓名</label>
        <div className="relative mt-1">
           <div className="w-full md:w-64 bg-slate-200 border border-slate-300 rounded-lg py-2 pl-9 pr-4 h-[42px] flex items-center animate-pulse">
                Loading...
           </div>
        </div>
      </div>
    );
  }

  return (
    <Combobox as="div" value={selectedAthlete} onChange={handleSelection} nullable>
      <Combobox.Label className="text-sm font-medium text-slate-600">運動員姓名</Combobox.Label>
      <div className="relative mt-1">
        <Combobox.Input
          className="w-full md:w-64 bg-white border border-slate-300 rounded-lg py-2 pl-9 pr-10 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition"
          onChange={handleQueryChange}
          displayValue={(athlete: string | null) => athlete || ''}
          autoComplete="off"
        />
        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
          <Search className="w-4 h-4 text-slate-400" />
        </div>
        <Combobox.Button className="absolute inset-y-0 right-0 flex items-center pr-2">
          <ChevronsUpDown className="h-5 w-5 text-gray-400" aria-hidden="true" />
        </Combobox.Button>

        {displayedAthletes && displayedAthletes.length > 0 && (
          <Combobox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
            {displayedAthletes.map((athlete, index) => (
              <Combobox.Option
                key={`${athlete}-${index}`}
                value={athlete}
                className={({ active }) =>
                  clsx(
                    'relative cursor-default select-none py-2 pl-10 pr-4',
                    active ? 'bg-blue-600 text-white' : 'text-gray-900'
                  )
                }
              >
                {({ selected, active }) => (
                  <>
                    <span className={clsx('block truncate', selected && 'font-medium')}>
                      {athlete}
                    </span>
                    {selected ? (
                      <span
                        className={clsx(
                          'absolute inset-y-0 left-0 flex items-center pl-3',
                          active ? 'text-white' : 'text-blue-600'
                        )}
                      >
                        <Check className="h-5 w-5" aria-hidden="true" />
                      </span>
                    ) : null}
                  </>
                )}
              </Combobox.Option>
            ))}
          </Combobox.Options>
        )}
      </div>
    </Combobox>
  );
};

export default AthleteSelector;
