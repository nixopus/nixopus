import * as React from 'react';
import { Check, Palette } from 'lucide-react';
import { useTheme } from 'next-themes';
import {
    CommandDialog,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
} from './command'; 
import { Button } from './button';
import { palette, themeColors } from '@/components/colors';

export function ModeToggler() {
    const [open, setOpen] = React.useState(false);
    const { setTheme, theme } = useTheme();
    const [prevTheme, setPrevTheme] = React.useState<string>(theme || 'light');

    React.useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === 'p' && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                setOpen((open) => !open);
            }
        };

        document.addEventListener('keydown', down);
        return () => document.removeEventListener('keydown', down);
    }, []);

    const handleOnFocus = (theme: string) => {
        setTheme(theme);
    };

    const handleSelected = (value: string) => {
        setPrevTheme(value);
        setTheme(value);
        setOpen(false);
    };

    const getThemeColors = (themeName: string) => {
        if (themeName)
            return [
                `hsl(${themeColors[themeName ?? 'light'].primary})`,
                `hsl(${themeColors[themeName ?? 'light'].secondary})`,
                `hsl(${themeColors[themeName ?? 'light'].background})`,
            ];
    };

    return (
        <>
            <div className="flex justify-center">
                <div className="relative">
                    <Button
                        variant={'outline'}
                        className="text-foreground no-underline hover:no-underline"
                        onClick={() => setOpen(true)}
                    >
                        <Palette className="h-4 w-4 text-primary" />
                    </Button>
                </div>
        </div>
            <CommandDialog open={open} onOpenChange={setOpen}>
                <CommandInput placeholder="Search for a theme" />
                <CommandList>
                    <CommandEmpty>No results found.</CommandEmpty>
                    <CommandGroup heading="Themes">
                        {palette.map((themeName) => (
                            <CommandItem
                                key={themeName}
                                onSelect={() => handleSelected(themeName)}
                                onMouseEnter={() => handleOnFocus(themeName)}
                                onMouseLeave={() => setTheme(prevTheme)}
                                className="flex cursor-pointer items-center justify-between"
                            >
                                <div className="flex items-center">
                                    {themeName === prevTheme && (
                                        <Check className="absolute mr-2 max-h-4 max-w-4" />
                                    )}
                                    <span className="pl-5 capitalize">{themeName}</span>
                                </div>
                                <div
                                    className="flex space-x-1 rounded-xl bg-muted px-2 py-1"
                                    style={{ background: 'themeName' }}
                                >
                                    {getThemeColors(themeName)?.map((color, index) => (
                                        <div
                                            key={index}
                                            className={`h-4 w-4 rounded-full`}
                                            style={{ backgroundColor: color }}
                                        ></div>
                                    ))}
                                </div>
                            </CommandItem>
                        ))}
                    </CommandGroup>
                </CommandList>
            </CommandDialog>
        </>
    );
}