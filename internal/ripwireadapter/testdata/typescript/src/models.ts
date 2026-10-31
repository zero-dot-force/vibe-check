/**
 * Model declarations exercising the ripwire adapter's TypeScript
 * classification scenarios. See FIXTURES.md at the testdata root for
 * the full fixture manifest.
 */

/** Top-level exported interface. */
export interface Point {
  x: number;
  y: number;
}

/** Top-level exported interface. */
export interface Named {
  name: string;
}

/** Top-level exported interface referencing another interface. */
export interface Repository {
  find(id: string): Point | undefined;
  save(point: Point): void;
}

/** Namespace hosting a nested exported interface. */
export namespace Shapes {
  /** Interface nested inside a namespace. */
  export interface Point {
    x: number;
    y: number;
  }
}

/**
 * Class whose four methods split into two field-access groups:
 * setX/readX touch only this.x and setY/readY touch only this.y, so
 * the methods form exactly two connected components (LCOM 2). The
 * property declarations are not methods and each touches a single
 * field, so they cannot merge the two components under any
 * interpretation.
 */
export class SplitMethods {
  x = 0;
  y = 0;

  setX(value: number): void {
    this.x = value;
  }

  readX(): number {
    return this.x;
  }

  setY(value: number): void {
    this.y = value;
  }

  readY(): number {
    return this.y;
  }
}
