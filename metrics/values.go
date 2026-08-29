package metrics

// Instability represents the instability metric I = Ce / (Ca + Ce).
// Value range: [0.0, 1.0] where 0.0 is maximally stable and 1.0 is maximally unstable.
// A maximally stable module (I = 0.0) has many dependents and no dependencies,
// making it costly to change. A maximally unstable module (I = 1.0) has no
// dependents and many dependencies, making it easy to change.
// When both Ca and Ce are 0, Instability is 0.0 (maximally stable by convention).
// Citation: Robert C. Martin, "Agile Software Development" (2003).
type Instability float64

// Abstractness represents the ratio of abstract types to total exported types.
// Value range: [0.0, 1.0] where 0.0 is fully concrete and 1.0 is fully abstract.
// An abstract type is a type that cannot be directly instantiated and serves
// as a contract for implementations (e.g., Go interfaces, Python ABCs).
// When a module has no exported types, Abstractness is 0.0.
// Citation: Robert C. Martin, "Agile Software Development" (2003).
type Abstractness float64

// Distance represents the Distance from Main Sequence metric D = |A + I - 1|.
// Value range: [0.0, 1.0] where 0.0 indicates the module lies on the main sequence
// (the ideal balance between abstractness and instability).
// D = 1.0 indicates the module is maximally far from the main sequence, either
// in the "zone of pain" (concrete and stable) or the "zone of uselessness"
// (abstract and unstable).
// Citation: Robert C. Martin, "Agile Software Development" (2003).
type Distance float64

// LCOM represents the Lack of Cohesion of Methods metric using the LCOM4 variant
// (Hitz & Montazeri, "Measuring Coupling and Cohesion in Object-Oriented Systems", 1995).
// LCOM4 counts connected components in the method-field graph, where methods are
// connected if they access at least one common field.
//
// Value semantics:
//   - LCOM = 0: no methods or fields (trivially cohesive)
//   - LCOM = 1: fully cohesive (all methods form a single connected component)
//   - LCOM > 1: can be split into LCOM independent classes
//
// The "fields" concept maps to language-specific shared state: struct fields in Go,
// instance attributes in Python, class properties in TypeScript. Each adapter
// documents its mapping.
//
// Limitation: LCOM4 does not account for method call chains — two methods that
// share no fields but call each other are treated as disconnected.
type LCOM int
