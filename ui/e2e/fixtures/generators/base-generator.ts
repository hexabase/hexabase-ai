import { faker } from '@faker-js/faker';

/**
 * Base class for all data generators
 * Provides common utilities and patterns
 */
export abstract class BaseGenerator<T> {
  protected faker = faker;
  
  constructor(protected seed?: number) {
    if (seed) {
      faker.seed(seed);
    }
  }
  
  /**
   * Generate a single instance
   */
  abstract generate(overrides?: Partial<T>): T;
  
  /**
   * Generate multiple instances
   */
  generateMany(count: number, overrides?: Partial<T>): T[] {
    return Array.from({ length: count }, () => this.generate(overrides));
  }
  
  /**
   * Generate with specific traits
   */
  abstract withTraits(traits: string[]): T;
  
  /**
   * Reset faker seed
   */
  resetSeed(seed?: number) {
    if (seed) {
      this.faker.seed(seed);
    } else {
      this.faker.seed();
    }
  }
  
  /**
   * Generate a unique ID with prefix
   */
  protected generateId(prefix: string): string {
    return `${prefix}-${this.faker.string.alphanumeric(8)}`;
  }
  
  /**
   * Generate a slug from name
   */
  protected generateSlug(name: string): string {
    return name.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '');
  }
  
  /**
   * Pick random items from array
   */
  protected pickRandom<I>(items: I[], count: number = 1): I[] {
    const shuffled = [...items].sort(() => 0.5 - Math.random());
    return shuffled.slice(0, count);
  }
  
  /**
   * Generate date within range
   */
  protected generateDateInRange(start: Date, end: Date): Date {
    return this.faker.date.between({ from: start, to: end });
  }
}

/**
 * Builder pattern for complex object creation
 */
export abstract class Builder<T> {
  protected data: Partial<T> = {};
  
  abstract build(): T;
  
  /**
   * Reset builder state
   */
  reset(): this {
    this.data = {};
    return this;
  }
  
  /**
   * Set a property value
   */
  protected set<K extends keyof T>(key: K, value: T[K]): this {
    this.data[key] = value;
    return this;
  }
}