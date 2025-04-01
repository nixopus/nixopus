"use client"
import MenuIcon from '../assets/icons/menu.svg';
import Link from 'next/link';
import { useState } from 'react';

export const Navbar = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  return (
    <div className="sticky top-0 z-50 bg-gradient-to-b from-gray-900 via-gray-950 to-black backdrop-blur-sm">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
      <div className="px-4 relative z-10">
        <div className="container">
          <div className="py-4 flex items-center justify-between">
            <Link href="/" className="relative group">
              <img 
                src="/nixopus_logo_transparent.png" 
                alt="Nixopus Logo" 
                className="h-24 w-24 relative mt-1 object-contain transition-transform duration-300 group-hover:scale-105"
              />
            </Link>
            
            <button 
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className='border border-white/20 h-10 w-10 inline-flex justify-center items-center rounded-lg sm:hidden hover:bg-gradient-to-r hover:from-indigo-500/10 hover:to-purple-500/10 transition-all duration-300'
              aria-label="Toggle menu"
            >
              <MenuIcon className="text-white" />
            </button>

            <nav className={`text-white gap-6 items-center ${isMenuOpen ? 'flex flex-col absolute top-full left-0 right-0 bg-gray-900/95 backdrop-blur-sm py-4' : 'hidden sm:flex'}`}>
              <Link 
                href="https://github.com/raghavyuva/nixopus" 
                target="_blank" 
                rel="noopener noreferrer" 
                className='text-opacity-60 text-white hover:text-opacity-100 hover:text-indigo-300 transition-all duration-300 px-4 py-2 rounded-lg hover:bg-white/5'
              >
                About
              </Link>
              <Link 
                href="https://docs.nixopus.com" 
                target="_blank" 
                rel="noopener noreferrer" 
                className='text-opacity-60 text-white hover:text-opacity-100 hover:text-purple-300 transition-all duration-300 px-4 py-2 rounded-lg hover:bg-white/5'
              >
                Features
              </Link>
              <Link 
                href="https://github.com/raghavyuva/nixopus/releases" 
                target="_blank" 
                rel="noopener noreferrer" 
                className='text-opacity-60 text-white hover:text-opacity-100 hover:text-pink-300 transition-all duration-300 px-4 py-2 rounded-lg hover:bg-white/5'
              >
                Changelogs
              </Link>
              <Link 
                href="https://docs.nixopus.com" 
                target="_blank" 
                rel="noopener noreferrer" 
                className='text-opacity-60 text-white hover:text-opacity-100 hover:text-indigo-300 transition-all duration-300 px-4 py-2 rounded-lg hover:bg-white/5'
              >
                Docs
              </Link>
              <Link 
                href="/try" 
                className='bg-gradient-to-r from-indigo-500 to-purple-500 py-2 px-6 rounded-lg text-white hover:from-indigo-600 hover:to-purple-600 transition-all duration-300 hover:shadow-lg hover:shadow-indigo-500/20'
              >
                Try now
              </Link>
            </nav>
          </div>
        </div>
      </div>
    </div>
  )
};
