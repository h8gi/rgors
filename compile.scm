(use matchable)

(define (tail? next)
  (eq? (car next) 'return))

(define (vm-compile x next)
  (cond [(symbol? x)
		 (list 'refer x next)]
		[(pair? x)
		 (match x
		   [('quote obj)
			(list 'constant obj next)]
		   [('lambda vars body)
			(list 'close vars (vm-compile body '(return)) next)]
		   [('if test then els)
			(let ([thenc (vm-compile then next)]
				  [elsec (vm-compile els next)])
			  (vm-compile test (list 'test thenc elsec)))]
		   [('set! var x)
			(vm-compile x (list 'assign var next))]
		   [('call/cc x)
			(let ([c (list 'conti
						   (list 'argument
								 (vm-compile x '(apply))))])
			  (if (tail? next)
				  c
				  (list 'frame next c)))]
		   [else
			(let loop ([args (cdr x)]
					   [c (vm-compile (car x) '(apply))])
			  (if (null? args)
				  (if (tail? next)
					  c
					  (list 'frame next c))
				  (loop (cdr args)
						(vm-compile (car args)
									(list 'argument c)))))])]
		[else
		 (list 'constant x next)]))

(define (lookup var e)
  (let nxtrib ([e e])
    (let nxtelt ([vars (caar e)]
                 [vals (cdar e)])
      (cond
       [(null? vars) (nxtrib (cdr e))]
       [(eq? (car vars) var) vals]
       [else (nxtelt (cdr vars) (cdr vals))]))))

(define (extend e vars vals)
  (cons (cons vars vals) e))

(define (closure body e vars)
  (list body e vars))

(define (continuation s)
  (closure (list 'nuate s 'v) '() '(v)))

(define (call-frame x e r s)
  (list x e r s))

;; (record (var ...) val exp ...) −→
;;     (apply (lambda (var ...) exp ...) val)

(define-syntax record
  (syntax-rules ()
	[(_ (var ...) val expr ...)
	 (apply (lambda (var ...) expr ...) val)]))

(define (VM a x e r s)
  (printf "a: ~A\nx: ~A\ne: ~A\nr: ~A\ns: ~A\n\n" a x e r s)
  (match x
    [('halt) a]
    [('refer var x)
     (VM (car (lookup var e)) x e r s)]
    [('constant obj x)
     (VM obj x e r s)]
    [('close vars body x)
     (VM (closure body e vars) x e r s)]
    [('test then else)
     (VM a (if a then else) e r s)]
    [('assign var x)
     (set-car! (lookup var e) a)
     (VM a x e r s)]
    [('conti x)
     (VM (continuation s) x e r s)]
    [('nuate s var)
     (VM (car (lookup var e)) '(return) e r s)]
    [('frame ret x)
     (VM a x e '() (call-frame ret e r s))]
    [('argument x)
     (VM a x e (cons a r) s)]
    [('apply)
     (record (body e vars) a
             (VM a body (extend e vars r) '() s))]
    [('return)
     (record (x e r s) s
             (VM a x e r s))]))

(define (vm-eval x)
  (VM '() (vm-compile x '(halt)) '() '() '()))
