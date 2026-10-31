package ripwireadapter

import "github.com/zero-dot-force/vibe-check/metrics"

// warningTSAbstractnessBias is the warning code documenting the two
// systematic biases in ripwire's TypeScript interface share.
const warningTSAbstractnessBias = "ts-abstractness-bias"

// tsAbstractnessBiasWarning returns the TypeScript abstractness bias
// warning. TypeScript abstractness is ripwire's interface share (design
// D3: exact for the fixture corpus, including interfaces nested in
// namespaces), with two documented biases: `abstract class` declarations
// are not counted as abstract types (undercount — conservative, abstract
// code reads more concrete than it is), and type aliases inflate the
// exported-type denominator (abstractness reads slightly lower than an
// interface-only reading would). Unlike Python — where ripwire's
// abstractness is structurally zero and vibe-check's own light parse is
// the source — TypeScript keeps ripwire's share because the spike
// validated it as exact on the fixture corpus.
func tsAbstractnessBiasWarning() metrics.Warning {
	return metrics.Warning{
		Code: warningTSAbstractnessBias,
		Message: "TypeScript abstractness is ripwire's interface share: abstract class " +
			"declarations are undercounted (conservative) and type aliases inflate the " +
			"exported-type denominator",
	}
}
