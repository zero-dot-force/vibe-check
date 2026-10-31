/**
 * Cycle side B of the two-directory import cycle: this module imports
 * from ../cycle-a, and ../cycle-a imports back from ../cycle-b.
 * Runtime execution is not required; the pair exercises static
 * import-graph cycle detection.
 */

import { markerA } from "../cycle-a";

/** Const imported by cycle-a, completing the cycle. */
export const markerB = "marker-b";

/** Consume the imported binding so the dependency edge is unambiguous. */
export function readMarkerA(): string {
  return markerA;
}
