import React from 'react';
import { TerminalIcon } from 'lucide-react';
import { cn } from './lib/utils';

function Terminal() {
  return (
    <div className={cn("group text-white  p-8 md:p-12 lg:p-16")}>
      <div className='container mx-auto'>
        <h1 className='text-4xl md:text-3xl font-bold flex items-center gap-4 font-mono text-yellow-400 mb-8'>
          <TerminalIcon className='w-12 h-12 animate-pulse' />
          <span className='bg-clip-text text-transparent bg-gradient-to-r to-red-400 from-yellow-400 '>
            Terminal
          </span>
        </h1>
        <h2 className='text-2xl md:text-xl font-semibold text-white mb-6 leading-tight'>
          Say goodbye to tedious page switching.
        </h2>
        <ul className='space-y-4 text-lg'>
          <li className='flex items-center'>
            <span className='text-yellow-400 mr-2'>►</span>
            Press <kbd className='bg-neutral-700 px-2 py-1 rounded mx-1'>Ctrl</kbd> + <kbd className='bg-neutral-700 px-2 py-1 rounded mx-1'>J</kbd> to open a persistent terminal
          </li>
          <li className='flex items-center'>
            <span className='text-yellow-400 mr-2'>►</span>
            Terminal sticks with you throughout all pages
          </li>
          <li className='flex items-center'>
            <span className='text-yellow-400 mr-2'>►</span>
            Customize layout: bottom or right, as per your preference
          </li>
        </ul>
      </div>
    </div>
  );
}

export default Terminal;