import React, { useState, useRef, useEffect } from 'react';
import { X } from 'lucide-react'; // Import X icon for close button

interface TooltipProps {
  children: React.ReactNode;
  text: string;
}

const Tooltip: React.FC<TooltipProps> = ({ children, text }) => {
  const [isVisible, setIsVisible] = useState(false);
  const tooltipRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLDivElement>(null); // Ref for the element that triggers the tooltip

  const toggleTooltip = (e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent event from bubbling up to parent elements
    setIsVisible(prev => !prev);
  };

  // Close tooltip when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        tooltipRef.current &&
        !tooltipRef.current.contains(event.target as Node) &&
        triggerRef.current &&
        !triggerRef.current.contains(event.target as Node)
      ) {
        setIsVisible(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [tooltipRef, triggerRef]);


  return (
    <div className="relative" ref={triggerRef}> {/* Removed flex items-center */}
      <span onClick={toggleTooltip} className="cursor-pointer"> {/* Make children clickable */}
        {children}
      </span>
      {isVisible && (
        <div
          ref={tooltipRef}
          className="absolute z-50 p-2 bg-slate-700 text-white text-xs rounded-md shadow-lg transition-opacity duration-300
                     max-w-xs w-max whitespace-normal break-words // Responsive width
                     bottom-full mb-2 // Default position above
                     left-0 // Align to the left of the trigger
                     "
          // Removed inline style transform: 'translateX(-50%)'
          // Dynamic positioning for mobile to prevent clipping, if more advanced logic is needed
          // For now, let's stick with bottom-full and left-0.
        >
          {text}
          <button
            onClick={toggleTooltip} // Click to dismiss
            className="absolute top-0 right-0 p-1 text-slate-300 hover:text-white"
            aria-label="Close tooltip"
          >
            <X size={12} />
          </button>
        </div>
      )}
    </div>
  );
};

export default Tooltip;
