
import LogoImage from '../assets/icons/logo.svg';
import MenuIcon from '../assets/icons/menu.svg';

export const Navbar = () => {
  return (
    <div className="bg-black">
      <div className="px-4">
        <div className="container bg-black">
          <div className="py-4 flex items-center justify-between">

            <div className="relative">
              <div className='absolute w-full top-2 bottom-0  blur-md '></div>

              <LogoImage className="h-24 w-24 relative mt-1 text-red-500" />
            </div>
            <div className='border border-white border-opacity-30 h-10 w-10 inline-flex justify-center items-center rounded-lg sm:hidden'>

              <MenuIcon className="text-white" />
            </div>
            <nav className='text-white gap-6 items-center hidden sm:flex'>

              <a href="#" className='text-opacity-60 text-white hover:text-opacity-100 transition' >About</a>
              <a href="#" className='text-opacity-60 text-white hover:text-opacity-100 transition'>Features</a>
              <a href="#" className='text-opacity-60 text-white hover:text-opacity-100 transition'>Sponsors</a>
              <a href="#" className='text-opacity-60 text-white hover:text-opacity-100 transition'>Changelogs</a>
              <a href="#" className='text-opacity-60 text-white hover:text-opacity-100 transition'>Docs</a>
              <button className='bg-white py-1 px-4 rounded-lg text-black'>Try now</button>
            </nav>
          </div>
        </div>
      </div>
    </div>
  )
};
