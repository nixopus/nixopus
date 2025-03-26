import React from 'react';
import { Package, GitBranch, Bell } from 'lucide-react';
import { cn } from './lib/utils';

function SelfHost() {
  return (
    <div className={cn("group text-white  p-8 md:p-12 lg:p-16")}>

      <div className='container mx-auto'>
        <h1 className='text-4xl md:text-3xl font-bold flex items-center gap-4 font-mono text-yellow-400 mb-8'>
          <Package className='w-12 h-12 animate-pulse' />
          <span className='bg-clip-text text-transparent bg-gradient-to-r to-red-400 from-yellow-400 '>
            Self-Host
          </span>
        </h1>
        <h2 className='text-xl md:text-xl font-semibold text-white mb-6 leading-tight'>
          Deploy and manage your GitHub projects with ease
        </h2>
        <ul className='space-y-6 text-lg'>
          <li className='flex items-start'>
            <Package className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
            <span>We&apos;ll build it for you whenever you push to GitHub.</span>
          </li>
          <li className='flex items-start'>
            <GitBranch className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
            <span>Get preview links for every pull request, making feature review a breeze.</span>
          </li>
          <li className='flex items-start'>
            <Bell className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
            <span>Monitor your application and receive timely notifications.</span>
          </li>
        </ul>
      </div>
    </div>
  );
}

export default SelfHost;