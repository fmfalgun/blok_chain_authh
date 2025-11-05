/**
 * Temperature Generator
 *
 * Generates realistic temperature readings using:
 * - Sine wave for day/night cycle
 * - Random noise for natural variation
 * - Configurable base temperature and amplitude
 */

class TemperatureGenerator {
    constructor(config) {
        this.baseTemp = config.baseTemp || 22;        // Base temperature (°C)
        this.amplitude = config.amplitude || 5;       // Temperature swing (±°C)
        this.noiseLevel = config.noiseLevel || 0.5;   // Random noise amplitude
        this.cycleHours = config.cycleHours || 24;    // Hours for complete cycle

        this.startTime = Date.now();

        console.log(`🌡️  Temperature Generator initialized:`);
        console.log(`   Base: ${this.baseTemp}°C`);
        console.log(`   Range: ${this.baseTemp - this.amplitude}°C to ${this.baseTemp + this.amplitude}°C`);
        console.log(`   Noise: ±${this.noiseLevel}°C`);
    }

    /**
     * Generate temperature reading
     *
     * Uses sine wave formula:
     * temp = baseTemp + amplitude * sin(2π * (currentHour / cycleHours))
     *
     * Day/night cycle simulation:
     * - Coldest at 6:00 AM (sin = -1)
     * - Warmest at 6:00 PM (sin = 1)
     */
    generate() {
        const now = Date.now();
        const elapsed = now - this.startTime;

        // Calculate current hour in the cycle (0-24)
        const currentHour = (elapsed / (1000 * 60 * 60)) % this.cycleHours;

        // Sine wave calculation (shifted so coldest is at hour 6)
        const angle = (2 * Math.PI * (currentHour - 6)) / this.cycleHours;
        const sineValue = Math.sin(angle);

        // Base temperature + sine wave + random noise
        const temperature = this.baseTemp +
                          (this.amplitude * sineValue) +
                          ((Math.random() - 0.5) * 2 * this.noiseLevel);

        // Round to 1 decimal place
        return Math.round(temperature * 10) / 10;
    }

    /**
     * Generate temperature for specific time of day
     * Useful for testing
     */
    generateForHour(hour) {
        const angle = (2 * Math.PI * (hour - 6)) / this.cycleHours;
        const sineValue = Math.sin(angle);

        const temperature = this.baseTemp +
                          (this.amplitude * sineValue) +
                          ((Math.random() - 0.5) * 2 * this.noiseLevel);

        return Math.round(temperature * 10) / 10;
    }

    /**
     * Get temperature statistics for full day
     */
    getDayStatistics() {
        const readings = [];
        for (let hour = 0; hour < 24; hour++) {
            readings.push(this.generateForHour(hour));
        }

        return {
            min: Math.min(...readings),
            max: Math.max(...readings),
            avg: readings.reduce((sum, t) => sum + t, 0) / readings.length,
            readings: readings
        };
    }
}

// Test if run directly
if (require.main === module) {
    console.log('\n🧪 Testing Temperature Generator\n');

    const generator = new TemperatureGenerator({
        baseTemp: 22,
        amplitude: 5,
        noiseLevel: 0.5,
        cycleHours: 24
    });

    console.log('📊 Sample readings:');
    for (let i = 0; i < 10; i++) {
        const temp = generator.generate();
        console.log(`  ${i + 1}. ${temp.toFixed(1)}°C`);
    }

    console.log('\n📈 24-hour statistics:');
    const stats = generator.getDayStatistics();
    console.log(`  Min: ${stats.min.toFixed(1)}°C`);
    console.log(`  Max: ${stats.max.toFixed(1)}°C`);
    console.log(`  Avg: ${stats.avg.toFixed(1)}°C`);

    console.log('\n🕐 Hourly breakdown:');
    for (let hour = 0; hour < 24; hour++) {
        const temp = generator.generateForHour(hour);
        console.log(`  ${String(hour).padStart(2, '0')}:00 - ${temp.toFixed(1)}°C`);
    }
}

module.exports = TemperatureGenerator;
