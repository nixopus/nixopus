import { Banner } from "@/components/Banner";
import { Navbar } from "@/components/Navbar";
// import { Sponsors } from "@/components/LogoTicker";
// import { ProductShowcase } from "@/components/ProductShowcase";
// import { FAQs } from "@/components/FAQs";
import { Footer } from "@/components/Footer";
import Hero from "@/components/Hero";
// import { FeaturesSectionDemo } from "@/components/Features";
import WaitlistForm from "@/components/WaitlistForm";

export default function Home() {
  return (
    <main className="min-h-screen bg-gradient-to-b from-gray-900 via-gray-950 to-black">
      <div className="relative">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
        
        <header className="relative z-10">
          {/* <Banner /> */}
          <Navbar />
        </header>

        <section id="hero" className="relative z-10">
          <Hero />
        </section>

        {/* Pre-release sections commented out */}
        {/* <section id="sponsors" aria-label="Our Sponsors">
          <Sponsors />
        </section>

        <section id="features" aria-label="Features">
          <FeaturesSectionDemo />
        </section>

        <section id="product" aria-label="Product Showcase">
          <ProductShowcase />
        </section>

        <section id="faq" aria-label="Frequently Asked Questions">
          <FAQs />
        </section> */}

        <section id="waitlist" className="relative z-10 py-16 sm:py-32">
          <WaitlistForm />
        </section>
      </div>

      <footer className="relative z-10">
        <Footer />
      </footer>
    </main>
  );
}
