# Gera testdata/eurachem.json e testdata/q_replicates.json, usados em
# robusto/eurachem_test.go, a partir dos 7 datasets do material de curso
# Eurachem "Hands-On Data Analytics" (ISO 13528:2022).
#
# Uso: Rscript testdata/eurachem.R <caminho/para/Rcode/0_function.R> testdata
# Requer os pacotes robustbase e moments.
#
# Algorithm A e média de Hampel são iterados até convergência estrita
# (1e-13 × s*), com fator 1.134 e início na mediana (ISO 13528:2022, C.3 e
# C.5.3), em vez de metRology::algA (maxiter = 30, fator γ exato) e rlm
# (início por mínimos quadrados) usados nos scripts do curso.
args <- commandArgs(TRUE)
fn <- readLines(args[1]); source(textConnection(fn[!grepl("install.packages", fn)]))
out_dir <- args[2]
suppressMessages({library(robustbase); library(moments)})

ds <- list()
ds$typical_PT <- c(
  0.04, 0.055, 0.178, 0.202, 0.206, 0.227, 0.228, 0.23,
  0.23, 0.235, 0.236, 0.237, 0.243, 0.244, 0.245, 0.2555,
  0.26, 0.264, 0.267, 0.27, 0.273, 0.274, 0.274, 0.278,
  0.2811, 0.287, 0.287, 0.288, 0.289, 0.295, 0.296, 0.311,
  0.331, 0.4246
)
ds$normal <- c(
  0.172, 0.181, 0.194, 0.201, 0.207, 0.213, 0.219, 0.224,
  0.229, 0.233, 0.237, 0.241, 0.244, 0.247, 0.250, 0.253,
  0.256, 0.259, 0.262, 0.265, 0.268, 0.271, 0.275, 0.279,
  0.283, 0.288, 0.293, 0.299, 0.305, 0.312, 0.321, 0.333,
  0.347, 0.365
)
ds$outliers_25 <- c(
  # Main distribution
  0.20, 0.21, 0.22, 0.23, 0.235, 0.24, 0.245, 0.25,
  0.255, 0.26, 0.265, 0.27, 0.275, 0.28, 0.285, 0.29,
  0.295, 0.30, 0.305, 0.31, 0.315, 0.32, 0.325, 0.33,
  0.335, 0.34,
  
  # Outliers
  0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.90
)
ds$outliers_50 <- c(
  # Main distribution
  0.20, 0.21, 0.22, 0.23, 0.235, 0.24, 0.245, 0.25,
  0.255, 0.26, 0.265, 0.27, 0.275, 0.28, 0.285, 0.29,
  
  # Outliers
  0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.85,
  0.90, 0.95, 1.00, 1.10, 1.20, 1.30, 1.40, 1.50
)
ds$skewed <- c(
  0.014687, 0.026498, 0.050597, 0.054278, 0.059851,
  0.071436, 0.073727, 0.092124, 0.101013, 0.127769,
  0.131322, 0.140219, 0.152077, 0.180326, 0.207713,
  0.209218, 0.233292, 0.263836, 0.279623, 0.293954,
  0.313394, 0.314573, 0.374857, 0.377487, 0.380751,
  0.387501, 0.396771, 0.403881, 0.414449, 0.436269,
  0.475793, 0.476479, 0.498918, 0.525854, 0.532951,
  0.607998, 0.632440, 0.643479, 0.692103, 0.700761,
  0.736669, 0.827701, 0.834227, 0.913490, 0.937803,
  0.955902, 1.024761, 1.203995, 1.268382, 1.357567
)
ds$skewed_outliers <- c(
  0, 0.01772065, 0.0618586, 0.06849186, 0.07592342,
  0.08425772, 0.09361322, 0.10412408, 0.11594187, 0.12923722,
  0.14420129, 0.15848234, 0.16104678, 0.1800081, 0.20134009,
  0.22531431, 0.25221162, 0.28230864, 0.31585507, 0.35303676,
  0.36238431, 0.39391736, 0.43834758, 0.48582633, 0.53167885,
  0.53529145, 0.58481027, 0.63113452, 0.63718698, 0.66909077,
  0.68508055, 0.69082447, 0.8, 1, 1.1, 1.2, 1.5, 2
)
ds$bimodal <- c(
  0.04, 0.055, 0.178, 0.202, 0.206, 0.227, 0.228, 0.23,
  0.23, 0.236, 0.244, 0.245, 0.267, 0.27, 0.274, 0.278,
  0.287, 0.288, 0.331, 0.5246, 0.535, 0.537, 0.543, 0.5555,
  0.56, 0.564, 0.573, 0.574, 0.5811, 0.587, 0.589, 0.595,
  0.596, 0.611
)

f <- function(v) sprintf("%.17g", v)
num <- function(v) paste0("[", paste(f(v), collapse = ","), "]")
alg_a <- function(x) {
  mu <- median(x); s <- mad(x)
  for (i in 1:10000) {
    z <- pmin(pmax(mu - 1.5 * s, x), mu + 1.5 * s)
    mu2 <- mean(z); s2 <- 1.134 * sd(z)
    done <- max(abs(mu2 - mu), abs(s2 - s)) <= 1e-13 * s2
    mu <- mu2; s <- s2
    if (done) break
  }
  c(mu, s)
}
hampel <- function(x, s) {
  m <- median(x)
  for (i in 1:10000) {
    w <- psi.hampel1((x - m) / s); m2 <- sum(w * x) / sum(w)
    if (abs(m2 - m) <= 1e-13 * s) return(m2)
    m <- m2
  }
  m
}
getmode <- function(v) { u <- unique(v); u[which.max(tabulate(match(v, u)))] }

cases <- c()
for (nm in names(ds)) {
  x <- ds[[nm]]
  a <- alg_a(x)
  qn <- s_Qn(x, mu.too = FALSE)
  q <- Qrobust_sd_between_labs(x, paste0("lab", seq_along(x)))$s_star
  cases <- c(cases, sprintf(paste0('{"name":"%s","input":%s,"mean":%s,"sd":%s,"median":%s,"mad":%s,',
    '"made":%s,"niqr":%s,"q1":%s,"q3":%s,"mode":%s,"skewness":%s,"alga_mean":%s,"alga_sd":%s,',
    '"qn":%s,"qn_hampel":%s,"q":%s,"q_hampel":%s,"made_hampel":%s}'),
    nm, num(x), f(mean(x)), f(sd(x)), f(median(x)), f(mad(x, constant = 1)), f(mad(x)),
    f(IQR(x) / 1.349), f(quantile(x, 0.25)), f(quantile(x, 0.75)), f(getmode(x)), f(skewness(x)),
    f(a[1]), f(a[2]), f(qn), f(hampel(x, qn)), f(q), f(hampel(x, q)), f(hampel(x, mad(x)))))
}
writeLines(c("[", paste(cases, collapse = ",\n"), "]"), file.path(out_dir, "eurachem.json"))

# Método Q com réplicas por laboratório.
rep_cases <- list(
  typical_PT_17x2 = list(x = ds$typical_PT, lab = rep(sprintf("lab%02d", 1:17), each = 2)),
  ties_6x2 = list(x = c(1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 5, 5), lab = rep(paste0("L", 1:6), each = 2)),
  unbalanced = list(x = c(10.1, 10.3, 9.8, 10.0, 10.2, 12.5, 9.9, 10.4, 10.0),
                    lab = c("A", "A", "B", "C", "C", "C", "D", "E", "E")),
  single_ties = list(x = c(5, 5, 5, 5, 6, 6, 6, 7, 8, 12), lab = paste0("L", 1:10))
)
rep_json <- sapply(names(rep_cases), function(nm) {
  r <- rep_cases[[nm]]
  sprintf('{"name":"%s","input":%s,"labs":[%s],"sd":%s}', nm, num(r$x),
          paste0('"', r$lab, '"', collapse = ","), f(Qrobust_sd_between_labs(r$x, r$lab)$s_star))
})
writeLines(c("[", paste(rep_json, collapse = ",\n"), "]"), file.path(out_dir, "q_replicates.json"))
