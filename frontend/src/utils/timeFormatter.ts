// frontend/src/utils/timeFormatter.ts
export const formatTime = (seconds: number | undefined | null, defaultValue: string = ''): string => {
  if (seconds === undefined || seconds === null) {
    return defaultValue;
  }
  if (seconds < 60) {
    return `${seconds.toFixed(2)}s`;
  }
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  const formattedSeconds = remainingSeconds.toFixed(2);
  const [secPart, msPart] = formattedSeconds.split('.');
  const paddedSecPart = secPart.padStart(2, '0');
  
  return `${minutes}:${paddedSecPart}.${msPart}`;
};