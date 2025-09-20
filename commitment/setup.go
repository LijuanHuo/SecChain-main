package commitment

import (
	"github.com/alinush/go-mcl"
)

func (pp *PublicParams) Paramgen(M int, N int) {
	pp.M = M
	pp.N = N
	pp.Pp_generators_alpha = make([]mcl.G1, 2*pp.M)
	pp.Pp_generators_alpha_beta = make([][]mcl.G1, pp.M)
	for i := 0; i < pp.M; i++ {
		pp.Pp_generators_alpha_beta[i] = make([]mcl.G1, 2*pp.N)
	}
	pp.Vp_generators_alpha = make([]mcl.G2, pp.M)
	pp.Vp_generators_beta = make([]mcl.G2, pp.N+1)

	//fmt.Println("Hello, World!")

	var alpha, beta, alpha_power, beta_power mcl.Fr

	alpha.Random()
	beta.Random()
	alpha_power.SetInt64(1)
	beta_power.SetInt64(1)

	//pp.G1.Random()
	//pp.G2.Random()

	pp.G1.HashAndMapTo(IntToBytes(1))
	pp.G2.HashAndMapTo(IntToBytes(1))

	for i := 0; i < pp.M; i++ {
		//fmt.Println("i=", i)

		mcl.FrMul(&alpha_power, &alpha_power, &alpha) //compute alpha^i
		mcl.G1Mul(&pp.Pp_generators_alpha[i], &pp.G1, &alpha_power)
		mcl.G2Mul(&pp.Vp_generators_alpha[i], &pp.G2, &alpha_power)

		alpha_beta_power := alpha_power
		for j := 0; j < pp.N; j++ {
			mcl.FrMul(&alpha_beta_power, &alpha_beta_power, &beta) //compute alpha^i beta^j
			mcl.G1Mul(&pp.Pp_generators_alpha_beta[i][j], &pp.G1, &alpha_beta_power)
		}
		//skip g1^{alpha^i beta^{n+1}}
		mcl.FrMul(&alpha_beta_power, &alpha_beta_power, &beta)
		pp.Pp_generators_alpha_beta[i][pp.N].SetString("0", 10)

		for j := pp.N + 1; j < 2*pp.N; j++ {
			mcl.FrMul(&alpha_beta_power, &alpha_beta_power, &beta) //compute alpha^i beta^j
			mcl.G1Mul(&pp.Pp_generators_alpha_beta[i][j], &pp.G1, &alpha_beta_power)
		}

		mcl.FrMul(&beta_power, &beta_power, &beta)
		mcl.G2Mul(&pp.Vp_generators_beta[i], &pp.G2, &beta_power)
	}

	//skip g1^{alpha^{n+1}}
	mcl.FrMul(&alpha_power, &alpha_power, &alpha)
	pp.Pp_generators_alpha[pp.M].SetString("0", 10)

	//compute g2^{beta^{n+1}}
	mcl.FrMul(&beta_power, &beta_power, &beta)
	mcl.G2Mul(&pp.Vp_generators_beta[pp.N], &pp.G2, &beta_power)

	for i := pp.M + 1; i < 2*pp.M; i++ {
		mcl.FrMul(&alpha_power, &alpha_power, &alpha) //compute alpha^i beta^j
		mcl.G1Mul(&pp.Pp_generators_alpha[i], &pp.G1, &alpha_power)
	}

	mcl.Pairing(&pp.Vp_gt_elt, &pp.Pp_generators_alpha[0], &pp.Vp_generators_alpha[pp.M-1])
}

// func (pp *PublicParams) Paramgen(N int) {
// 	pp.N = N
// 	pp.Pp_generators_alpha = make([]mcl.G1, 2*pp.N)
// 	pp.Pp_generators_alpha_beta = make([][]mcl.G1, pp.N)
// 	for i := 0; i < pp.N; i++ {
// 		pp.Pp_generators_alpha_beta[i] = make([]mcl.G1, 2*pp.N)
// 	}
// 	pp.Vp_generators_alpha = make([]mcl.G2, pp.N)
// 	pp.Vp_generators_beta = make([]mcl.G2, pp.N+1)

// 	//fmt.Println("Hello, World!")

// 	var alpha, beta, alpha_power, beta_power mcl.Fr

// 	alpha.Random()
// 	beta.Random()
// 	alpha_power.SetInt64(1)
// 	beta_power.SetInt64(1)

// 	//pp.G1.Random()
// 	//pp.G2.Random()

// 	pp.G1.HashAndMapTo(IntToBytes(1))
// 	pp.G2.HashAndMapTo(IntToBytes(1))

// 	for i := 0; i < pp.N; i++ {
// 		//fmt.Println("i=", i)

// 		mcl.FrMul(&alpha_power, &alpha_power, &alpha) //compute alpha^i
// 		mcl.G1Mul(&pp.Pp_generators_alpha[i], &pp.G1, &alpha_power)
// 		mcl.G2Mul(&pp.Vp_generators_alpha[i], &pp.G2, &alpha_power)

// 		alpha_beta_power := alpha_power
// 		for j := 0; j < pp.N; j++ {
// 			mcl.FrMul(&alpha_beta_power, &alpha_beta_power, &beta) //compute alpha^i beta^j
// 			mcl.G1Mul(&pp.Pp_generators_alpha_beta[i][j], &pp.G1, &alpha_beta_power)
// 		}
// 		//skip g1^{alpha^i beta^{n+1}}
// 		mcl.FrMul(&alpha_beta_power, &alpha_beta_power, &beta)
// 		pp.Pp_generators_alpha_beta[i][pp.N].SetString("0", 10)

// 		for j := pp.N + 1; j < 2*pp.N; j++ {
// 			mcl.FrMul(&alpha_beta_power, &alpha_beta_power, &beta) //compute alpha^i beta^j
// 			mcl.G1Mul(&pp.Pp_generators_alpha_beta[i][j], &pp.G1, &alpha_beta_power)
// 		}

// 		mcl.FrMul(&beta_power, &beta_power, &beta)
// 		mcl.G2Mul(&pp.Vp_generators_beta[i], &pp.G2, &beta_power)
// 	}

// 	//skip g1^{alpha^{n+1}}
// 	mcl.FrMul(&alpha_power, &alpha_power, &alpha)
// 	pp.Pp_generators_alpha[pp.N].SetString("0", 10)

// 	//compute g2^{beta^{n+1}}
// 	mcl.FrMul(&beta_power, &beta_power, &beta)
// 	mcl.G2Mul(&pp.Vp_generators_beta[pp.N], &pp.G2, &beta_power)

// 	for i := pp.N + 1; i < 2*pp.N; i++ {
// 		mcl.FrMul(&alpha_power, &alpha_power, &alpha) //compute alpha^i beta^j
// 		mcl.G1Mul(&pp.Pp_generators_alpha[i], &pp.G1, &alpha_power)
// 	}

// 	mcl.Pairing(&pp.Vp_gt_elt, &pp.Pp_generators_alpha[0], &pp.Vp_generators_alpha[pp.N-1])
// }
