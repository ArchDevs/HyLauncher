import bgServersImage from "../assets/images/bg-servers.png";

function BackgroundServers() {
  return (
    <div
      className="absolute inset-0 bg-cover bg-center z-0 scale-105"
      style={{
        backgroundImage: `url(${bgServersImage})`,
      }}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-transparent to-[#090909]/[0.50]" />
    </div>
  );
}

export default BackgroundServers;
