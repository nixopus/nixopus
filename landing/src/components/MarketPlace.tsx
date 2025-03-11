import React from 'react';
import {  Download, Trash2, Activity, Store } from 'lucide-react';
import { cn } from './lib/utils';

function MarketPlace() {
    return (
        <div className={cn("group text-white  p-8 md:p-12 lg:p-16")}>
            <div className='container mx-auto'>
                <h1 className='text-xl md:text-3xl font-bold flex items-center gap-4 font-mono mb-8'>
                    <Store className='w-12 h-12 animate-pulse text-yellow-400' />
                    <span className='bg-clip-text text-transparent bg-gradient-to-r from-yellow-400 to-red-400'>
                        Market Place
                    </span>
                </h1>
                <h2 className='text-xl md:text-xl font-semibold text-white mb-6 leading-tight'>
                    Host 100+ open source applications on your server with a single click
                </h2>
                <ul className='space-y-6 text-lg'>
                    <li className='flex items-start'>
                        <Download className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
                        <span>Install applications and get instant hosted links</span>
                    </li>
                    <li className='flex items-start'>
                        <Trash2 className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
                        <span>Uninstall at any time with ease</span>
                    </li>
                    <li className='flex items-start'>
                        <Activity className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
                        <span>Monitor and view hosted app logs effortlessly</span>
                    </li>
                </ul>
            </div>
        </div>
    );
}

export default MarketPlace;