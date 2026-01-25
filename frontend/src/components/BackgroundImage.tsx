function BackgroundImage() {
  return (
    <div
      className="absolute inset-0 bg-cover bg-center z-0 scale-105"
      style={{
        backgroundImage: `url('https://hytale.com/static/images/backgrounds/content-upper-new-1920.jpg')`,
      }}
    >
      {/* Darkening overlay so the UI stays readable */}
      <div className="absolute inset-0 bg-black/40" />
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-[#090909]" />
    </div>
  );
}

export default BackgroundImage;
