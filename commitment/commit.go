package commitment

import "github.com/alinush/go-mcl"

// / generate a new global commitment.
func (com *GlobalCommitment) NewGlocalCommitment(pp PublicParams, values []mcl.Fr) {
	rearrange_alpha_beta_vec := make([]mcl.G1, pp.N*pp.N)
	for i := 0; i < pp.N; i++ {
		for j := 0; j < pp.N; j++ {
			rearrange_alpha_beta_vec[i*pp.N+j] = pp.Pp_generators_alpha_beta[j][i]
		}
	}

	//global_commitment=g_1^{\sum_{i\in[N]}\sum_{j\in[N]}m_{ij}\alpha^j\beta^i}
	mcl.G1MulVec(&com.Global_commitment, rearrange_alpha_beta_vec, values)
}

// / generate a new local commitment.
func (com *LocalCommitment) NewLocalCommitment(pp PublicParams, values []mcl.Fr) {
	//com.local_commitment = make([]mcl.G1, pp.N)
	//for i := 0; i < pp.N; i++ {
	mcl.G1MulVec(&com.Local_commitment, pp.Pp_generators_alpha[0:pp.N], values)
	//}
}

// / updated an existing global commitment
func (com *GlobalCommitment) UpdateGlobalCommitment(
	pp PublicParams,
	changedIndex []int,
	deltaValue mcl.Fr) {
	var res_c2 mcl.G1
	mcl.G1Mul(&res_c2, &pp.Pp_generators_alpha_beta[changedIndex[0]][changedIndex[1]], &deltaValue)
	mcl.G1Add(&com.Global_commitment, &com.Global_commitment, &res_c2)
}

// / updated an existing local commitment
func (com *LocalCommitment) UpdateLocalCommitment(
	pp PublicParams,
	changedIndex []int,
	deltaValue mcl.Fr) {
	var resci mcl.G1
	mcl.G1Mul(&resci, &pp.Pp_generators_alpha[changedIndex[1]], &deltaValue)
	mcl.G1Add(&com.Local_commitment, &com.Local_commitment, &resci)
}
