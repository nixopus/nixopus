import { useState, useEffect } from 'react';

interface WeatherData {
  temperature: number;
  condition: string;
  description: string;
  humidity: number;
  windSpeed: number;
  location: string;
}

interface UseWeatherProps {
  latitude?: number;
  longitude?: number;
  location?: string;
}

const useWeather = ({ latitude, longitude, location = 'Default Location' }: UseWeatherProps) => {
  const [weather, setWeather] = useState<WeatherData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [userLocation, setUserLocation] = useState<{ lat: number; lon: number } | null>(null);

  useEffect(() => {
    if (!latitude || !longitude) {
      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
          (position) => {
            setUserLocation({
              lat: position.coords.latitude,
              lon: position.coords.longitude
            });
          },
          () => {
            setUserLocation({ lat: 40.7128, lon: -74.0060 });
          }
        );
      } else {
        setUserLocation({ lat: 40.7128, lon: -74.0060 });
      }
    }
  }, [latitude, longitude]);

  useEffect(() => {
    const fetchWeather = async () => {
      try {
        setIsLoading(true);
        setError(null);

        const locationToUse = userLocation || {
          lat: latitude ?? 40.7128,
          lon: longitude ?? -74.0060
        };

        const apiKey = process.env.NEXT_PUBLIC_WEATHER_API_KEY;
        
        if (!apiKey) {
          const mockData: WeatherData = {
            temperature: Math.round(Math.random() * 30 + 10),
            condition: 'Clear',
            description: 'Clear sky',
            humidity: Math.round(Math.random() * 40 + 40),
            windSpeed: Math.round(Math.random() * 10 + 5),
            location: location
          };
          setWeather(mockData);
          setIsLoading(false);
          return;
        }

        const response = await fetch(
          `https://api.openweathermap.org/data/2.5/weather?lat=${locationToUse.lat}&lon=${locationToUse.lon}&appid=${apiKey}&units=imperial`
        );

        if (!response.ok) {
          throw new Error('Failed to fetch weather data');
        }

        const data = await response.json();
        
        const weatherData: WeatherData = {
          temperature: Math.round(data.main.temp),
          condition: data.weather[0].main,
          description: data.weather[0].description,
          humidity: data.main.humidity,
          windSpeed: Math.round(data.wind.speed),
          location: data.name || location
        };

        setWeather(weatherData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load weather');
        const mockData: WeatherData = {
          temperature: 72,
          condition: 'Clear',
          description: 'Clear sky',
          humidity: 55,
          windSpeed: 7,
          location: location
        };
        setWeather(mockData);
      } finally {
        setIsLoading(false);
      }
    };

    fetchWeather();

    const interval = setInterval(fetchWeather, 300000);

    return () => clearInterval(interval);
  }, [latitude, longitude, location, userLocation]);

  return {
    weather,
    isLoading,
    error
  };
};

export default useWeather;
export type { WeatherData };

