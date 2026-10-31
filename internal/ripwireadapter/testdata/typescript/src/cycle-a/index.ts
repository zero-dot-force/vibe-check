/**
 * Cycle side A of the two-directory import cycle: this module imports
 * from ../cycle-b, and ../cycle-b imports back from ../cycle-a.
 * Runtime execution is not required; the pair exercises static
 * import-graph cycle detection.
 */

import { markerB } from "../cycle-b";

/** Const imported by cycle-b, completing the cycle. */
export const markerA = "marker-a";

/** Consume the imported binding so the dependency edge is unambiguous. */
export function readMarkerB(): string {
  return markerB;
}
