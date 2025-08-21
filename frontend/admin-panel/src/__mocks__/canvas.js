// Mock canvas module to avoid native dependency issues in tests
module.exports = {
  createCanvas: () => ({
    getContext: () => ({
      fillRect: jest.fn(),
      clearRect: jest.fn(),
      getImageData: jest.fn(() => ({ data: [] })),
      putImageData: jest.fn(),
      createImageData: jest.fn(() => ({ data: [] })),
      setTransform: jest.fn(),
      drawImage: jest.fn(),
      save: jest.fn(),
      restore: jest.fn(),
      beginPath: jest.fn(),
      moveTo: jest.fn(),
      lineTo: jest.fn(),
      closePath: jest.fn(),
      stroke: jest.fn(),
      fill: jest.fn(),
    }),
    toBuffer: jest.fn(),
    toDataURL: jest.fn(),
  }),
  loadImage: jest.fn(() => Promise.resolve({
    width: 100,
    height: 100
  })),
  registerFont: jest.fn(),
};